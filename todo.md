* Pass cleanup func to Start  
  `startStopper.Start(readyCh chan error, cleanup func() error)`  
  then there is no need to call `done(lastError)` implicitly  
  it must be called explicitly in `startstopper.Stop`  
  
  ```go
  func (srv *Srv) cleanup() error {
  	return nil
  }

  func (srv *Srv) Start(signal chan error) {
  	var lastError error

  	closingCh, err := srv.StartStopper.Start(signal, srv.cleanup)
  	if err != nil {
  		return
  	}

  	for {
  		select {
  		// listen to closingCh in main loop
  		// returns error from srv.cleanup
  		case err := <-closingCh:
  			return err

  		//	other cases
  		case v := <-srv.C:
  			fmt.Println(v)
  		}
  	}
  }
  ```
