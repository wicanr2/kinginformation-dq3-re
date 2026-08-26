package game

import (
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/config"
)

type recordingAudio struct {
	tracks  []int
	enabled []bool
}

func (a *recordingAudio) Play(track int)                       { a.tracks = append(a.tracks, track) }
func (*recordingAudio) PlaySFX(int)                            {}
func (*recordingAudio) SFXDurationNanos(int) int64             { return 0 }
func (a *recordingAudio) SetEnabled(on bool)                   { a.enabled = append(a.enabled, on) }
func (*recordingAudio) SetVolume(int)                          {}
func (*recordingAudio) SetSFXWithDurations([][]int16, []int64) {}
func (*recordingAudio) SetMBG([]byte)                          {}

func TestCurrentAudioCueFollowsVisibleScene(t *testing.T) {
	tests := []struct {
		name string
		game Game
		want string
	}{
		{name: "title", game: Game{showTitle: true}, want: audioCueTitle},
		{name: "field", game: Game{}, want: audioCueField},
		{name: "castle", game: Game{inTown: true, curCty: 0}, want: audioCueCastle},
		{name: "town", game: Game{inTown: true, curCty: 1}, want: audioCueTown},
		{name: "dungeon", game: Game{inTown: true, curCty: 7}, want: audioCueDungeon},
		{name: "battle", game: Game{inTown: true, curCty: 7, battle: Battle{active: true}}, want: audioCueBattle},
		{name: "ending", game: Game{battle: Battle{active: true}, lotoBlessed: true}, want: audioCueEnding},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.game.currentAudioCue(); got != tt.want {
				t.Fatalf("currentAudioCue()=%q，預期 %q", got, tt.want)
			}
		})
	}
}

func TestResumeCurrentMusicUsesPackTrack(t *testing.T) {
	audio := &recordingAudio{}
	g := &Game{pack: loadTestPack(t), music: audio, inTown: true, curCty: 7}
	g.resumeCurrentMusic()
	if len(audio.tracks) != 1 || audio.tracks[0] != 3 {
		t.Fatalf("迷宮恢復軌=%v，預期 pack dungeon track 3", audio.tracks)
	}
}

func TestSettingsReenableImmediatelyResumesCurrentScene(t *testing.T) {
	audio := &recordingAudio{}
	g := &Game{cfg: config.Default(), pack: loadTestPack(t), music: audio, inTown: true, curCty: 0}
	g.applySettingChange(setRowMusic, 1)
	g.applySettingChange(setRowMusic, 1)
	if len(audio.enabled) != 2 || audio.enabled[0] || !audio.enabled[1] {
		t.Fatalf("音樂開關呼叫=%v，預期 false→true", audio.enabled)
	}
	if len(audio.tracks) != 1 || audio.tracks[0] != 1 {
		t.Fatalf("重新啟用後播放=%v，預期 pack castle track 1", audio.tracks)
	}
}
