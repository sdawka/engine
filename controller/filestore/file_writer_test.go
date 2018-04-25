package filestore

import "testing"

type mockWriter struct {
	text string
	err  error
}

func (w *mockWriter) WriteString(s string) (int, error) {
	if w.err != nil {
		return 0, w.err
	}

	w.text += s
	return len(s), nil
}

func TestWriteGameInfo(t *testing.T) {

}
