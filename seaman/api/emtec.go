package api

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// Action IDs
	ActIdEmtec_SceneNext = "emtec_scenenext"
)

type Track struct {
	Id   int32
	Name string
}

func NewTrack(str string) (Track, error) {
	s := strings.Split(str, "__")
	if len(s) != 2 {
		return Track{}, fmt.Errorf("callbackValue (%s) is not expected", str)
	}
	trackId, err := strconv.Atoi(s[0])
	if err != nil {
		return Track{}, fmt.Errorf("callbackValue (%s) is not expected", str)
	}
	return Track{int32(trackId), s[1]}, nil
}

func (m Track) String() string {
	return fmt.Sprintf("%d__%s", m.Id, m.Name)
}
