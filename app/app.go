package app

// run - main run method
func Run() error {

	// check params
	if err := checkParams(); err != nil {
		return err
	}

	return nil
}
