package cmd

import "testing"

func TestValidateVideoOptions(t *testing.T) {
	if err := validateVideoOptions(121, 24, "", 0); err != nil {
		t.Fatal(err)
	}
	if err := validateVideoOptions(122, 24, "", 0); err == nil {
		t.Fatal("expected invalid num_frames error")
	}
	if err := validateVideoOptions(121, 61, "", 0); err == nil {
		t.Fatal("expected invalid frame_rate error")
	}
	if err := validateVideoOptions(121, 24, "keyframes", 1); err == nil {
		t.Fatal("expected keyframe image count error")
	}
}
