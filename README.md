# Battlesnake Engine

API and game logic for Battlesnake.

---

# Install and setup

1. Install Golang if you haven't already [here](https://golang.org/doc/install)

2. Set your GOPATH, [here](https://github.com/golang/go/wiki/SettingGOPATH)

3. At least for my setup, I had to add a path to my Go _bin_ folder. You can use:
`export PATH="$GOPATH/bin:$PATH"`

4. Once you've setup the engine project (mine is in `~/go/src/github.com/battlesnakeio/engine`). The docs for GOPATH and project directory layouts can be found [here](çhttps://github.com/golang/go/wiki/SettingGOPATH).

5. run `go get` to from the project folder to download your dependencies. 


# Running the engine

1. Build an excutable via `go build` (add a `-a` to rebuild without cache), then call `./engine` to run the newly created executable. **Better**: run `Make run`

Note: if you use the Makefile, you'll want JQ installed, [here](https://stedolan.github.io/jq/download/)

## Running a game with the CLI

1. Setup a `snake-config.json`, by default the engine looks in the HOME directory (see Make run-game in the Makefile)

Here's an example: 

```
{
    "width": 20,
    "height": 20,
    "food": 10,
    "snakes": [
        {
            "name": "Snake 1",
            "url": "http://localhost:8080"
        },
        {
            "name": "Snake 2",
            "url": "http://localhost:3001"
        }
    ]
}
```

2. Start the engine (refer above)

3. Start a game with `Make run-game`

Example:
```
⇒  Make run-game
go install github.com/battlesnakeio/engine/cmd/engine-cli
engine-cli run -g "d151fe9d-8c15-4d31-a932-e7b4248d5586"
```

3. To repeat the game run: `engine-cli replay -g <game id>`


---

# API 

Last checked April 16, 2018

**`/move`**

```
{
   "game":{
      "id":"580d0aee-e985-4383-a461-c84f88605d04"
   },
   "turn":5,
   "board":{
      "height":20,
      "width":20,
      "food":[
         {
            "x":13,
            "y":6
         },
         {
            "x":16,
            "y":19
         }
      ],
      "snakes":[
         {
            "id":"28556fdc-7bc8-4545-91f8-5ddfa347f1b7",
            "name":"Snake 1",
            "health":95,
            "body":[
               {
                  "x":15,
                  "y":0
               },
               {
                  "x":15,
                  "y":1
               },
               {
                  "x":15,
                  "y":2
               }
            ]
         },
         {
            "id":"8b49c87f-e0ab-443b-b8eb-f214a9ce58cf",
            "name":"Snake 2",
            "health":95,
            "body":[
               {
                  "x":5,
                  "y":1
               },
               {
                  "x":5,
                  "y":2
               },
               {
                  "x":5,
                  "y":3
               }
            ]
         }
      ]
   },
   "you":{
      "id":"28556fdc-7bc8-4545-91f8-5ddfa347f1b7",
      "name":"Snake 1",
      "health":95,
      "body":[
         {
            "x":15,
            "y":0
         },
         {
            "x":15,
            "y":1
         },
         {
            "x":15,
            "y":2
         }
      ]
   }
}
```
