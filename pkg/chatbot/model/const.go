package model

// Action IDs
const (
	// common
	ActIdCommon_Cancel = "common_cancel"
	// release
	ActIdRelease_SelectedRepository = "release_selected_repo"
	ActIdRelease_SelectedLevelMajor = "release_selected_level_major"
	ActIdRelease_SelectedLevelMinor = "release_selected_level_minor"
	ActIdRelease_SelectedLevelPatch = "release_selected_level_patch"
	ActIdRelease_OK                 = "release_ok"
)

// Callback Values
const (
	CallbackValueRelease_VersionMajor = "release/major"
	CallbackValueRelease_VersionMinor = "release/minor"
	CallbackValueRelease_VersionPatch = "release/patch"
)
