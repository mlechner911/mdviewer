package main

// singleSelection is a Wails open or save dialog.
type singleSelection interface {
	PromptForSingleSelection() (string, error)
}

// dialogCancelled is what Wails reports on Windows when the user cancels a
// file dialog. The error lives in an internal Wails package, so only its text
// can be compared. On macOS and Linux a cancelled dialog returns an empty path.
const dialogCancelled = "cancelled by user"

// promptPath shows dlg and returns the chosen path, "" when the user cancelled:
// cancelling is the same answer on every platform and never an error.
func promptPath(dlg singleSelection) (string, error) {
	path, err := dlg.PromptForSingleSelection()
	if err != nil {
		if err.Error() == dialogCancelled {
			return "", nil
		}
		return "", err
	}
	return path, nil
}
