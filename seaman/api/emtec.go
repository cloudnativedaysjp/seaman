package api

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// Action IDs
	ActIdEmtec_SceneNext = "emtec_scenenext"
	ActIdEmtec_OnAirNext = "emtec_onairnext"
)

type Track struct {
	Id         int32
	Name       string
	NextTalkId int32
}

func NewTrack(id int32, name string, nextTalkId int32) Track {
	return Track{id, name, nextTalkId}
}

func CastToTrack(str string) (Track, error) {
	s := strings.Split(str, "__")
	if len(s) != 3 {
		return Track{}, fmt.Errorf("callbackValue (%s) is not expected", str)
	}
	trackId, err := strconv.Atoi(s[0])
	if err != nil {
		return Track{}, fmt.Errorf("callbackValue (%s) is not expected", str)
	}
	nextTalkId, err := strconv.Atoi(s[2])
	if err != nil {
		return Track{}, fmt.Errorf("callbackValue (%s) is not expected", str)
	}
	return Track{int32(trackId), s[1], int32(nextTalkId)}, nil
}

func (m Track) String() string {
	return fmt.Sprintf("%d__%s__%d", m.Id, m.Name, m.NextTalkId)
}
