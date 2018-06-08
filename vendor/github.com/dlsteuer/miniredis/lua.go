package miniredis

import (
	"github.com/dlsteuer/miniredis/server"
	"github.com/go-redis/redis"
	"github.com/yuin/gopher-lua"
)

func setGlobalsProtection(l *lua.LState) error {
	// redis doesn't allow accidental global creation in the lua scripts
	// to mimic that we need to disable global creation in our lua scripting state
	// taken from https://github.com/antirez/redis/blob/unstable/src/scripting.c#L860
	return l.DoString(`local dbg=debug
local mt = {}
setmetatable(_G, mt)
mt.__newindex = function (t, n, v)
  print("new index")
  print(dbg.getinfo(2, "S").what)
  if dbg.getinfo(2) then
    error("Script attempted to create global variable '"..tostring(n).."'", 2)
  end
  rawset(t, n, v)
end
mt.__index = function (t, n)
  if dbg.getinfo(2) then
    error("Script attempted to access nonexistent global variable '"..tostring(n).."'", 2)
  end
  return rawget(t, n)
end
debug = nil`)
}

func mkLuaFuncs(conn *redis.Client) map[string]lua.LGFunction {
	mkCall := func(failFast bool) func(l *lua.LState) int {
		return func(l *lua.LState) int {
			top := l.GetTop()
			if top == 0 {
				l.Error(lua.LString("Please specify at least one argument for redis.call()"), 1)
				return 0
			}
			var args []interface{}
			for i := 1; i <= top; i++ {
				switch a := l.Get(i).(type) {
				// case lua.LBool:
				// args[i-2] = a
				case lua.LNumber:
					// value, _ := strconv.ParseFloat(lua.LVAsString(arg), 64)
					args = append(args, float64(a))
				case lua.LString:
					args = append(args, string(a))
				default:
					l.Error(lua.LString("Lua redis() command arguments must be strings or integers"), 1)
					return 0
				}
			}
			_, ok := args[0].(string)
			if !ok {
				l.Error(lua.LString("Unknown Redis command called from Lua script"), 1)
				return 0
			}
			c := redis.NewCmd(args...)
			err := conn.Process(c)
			if err != nil {
				if failFast {
					// call() mode
					l.Error(lua.LString(err.Error()), 1)
					return 0
				}
				// pcall() mode
				l.Push(lua.LNil)
				return 1
			}
			res, err := c.Result()
			if err != nil {
				if failFast {
					// call() mode
					l.Error(lua.LString(err.Error()), 1)
					return 0
				}
				// pcall() mode
				l.Push(lua.LNil)
				return 1
			}

			if res == nil {
				l.Push(lua.LNil)
			} else {
				switch r := res.(type) {
				case int64:
					l.Push(lua.LNumber(r))
				case []uint8:
					l.Push(lua.LString(string(r)))
				case []interface{}:
					l.Push(redisToLua(l, r))
				case string:
					l.Push(lua.LString(r))
				default:
					panic("type not handled")
				}
			}
			return 1
		}
	}

	return map[string]lua.LGFunction{
		"call":  mkCall(true),
		"pcall": mkCall(false),
		"error_reply": func(l *lua.LState) int {
			msg := l.CheckString(1)
			res := &lua.LTable{}
			res.RawSetString("err", lua.LString(msg))
			l.Push(res)
			return 1
		},
		"status_reply": func(l *lua.LState) int {
			msg := l.CheckString(1)
			res := &lua.LTable{}
			res.RawSetString("ok", lua.LString(msg))
			l.Push(res)
			return 1
		},
		"sha1hex": func(l *lua.LState) int {
			top := l.GetTop()
			if top != 1 {
				l.Error(lua.LString("wrong number of arguments"), 1)
				return 0
			}
			msg := lua.LVAsString(l.Get(1))
			l.Push(lua.LString(sha1Hex(msg)))
			return 1
		},
	}
}

func luaToRedis(l *lua.LState, c *server.Peer, value lua.LValue) {
	if value == nil {
		c.WriteNull()
		return
	}

	switch t := value.(type) {
	case *lua.LNilType:
		c.WriteNull()
	case lua.LBool:
		if lua.LVAsBool(value) {
			c.WriteInt(1)
		} else {
			c.WriteInt(0)
		}
	case lua.LNumber:
		c.WriteInt(int(lua.LVAsNumber(value)))
	case lua.LString:
		s := lua.LVAsString(value)
		if s == "OK" {
			c.WriteInline(s)
		} else {
			c.WriteBulk(s)
		}
	case *lua.LTable:
		// special case for tables with an 'err' or 'ok' field
		// note: according to the docs this only counts when 'err' or 'ok' is
		// the only field.
		if s := t.RawGetString("err"); s.Type() != lua.LTNil {
			c.WriteError(s.String())
			return
		}
		if s := t.RawGetString("ok"); s.Type() != lua.LTNil {
			c.WriteInline(s.String())
			return
		}

		result := []lua.LValue{}
		for j := 1; true; j++ {
			val := l.GetTable(value, lua.LNumber(j))
			if val == nil {
				result = append(result, val)
				continue
			}

			if val.Type() == lua.LTNil {
				break
			}

			result = append(result, val)
		}

		c.WriteLen(len(result))
		for _, r := range result {
			luaToRedis(l, c, r)
		}
	default:
		panic("....")
	}
}

func redisToLua(l *lua.LState, res []interface{}) *lua.LTable {
	rettb := l.NewTable()
	for _, e := range res {
		var v lua.LValue
		if e == nil {
			v = lua.LValue(nil)
		} else {
			switch et := e.(type) {
			case int64:
				v = lua.LNumber(et)
			case []uint8:
				v = lua.LString(string(et))
			case []interface{}:
				v = redisToLua(l, et)
			case string:
				v = lua.LString(et)
			default:
				// TODO: oops?
				v = lua.LString(e.(string))
			}
		}
		l.RawSet(rettb, lua.LNumber(rettb.Len()+1), v)
	}
	return rettb
}
