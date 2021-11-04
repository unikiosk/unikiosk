package file

import (
	"os"
)

// Exist checks to see if a file exist at the provided path.
func Exist(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// ignore missing files
			return false, nil
		}
		return false, err
	}
	defer f.Close()
	return true, nil
}
