package deployer

// close all tcp and socket connections
func (d *Deployer) Close() {
	// wait until current tasks are done
	d.waitGroup.Wait()

	d.ListenerSocket.Close()
	d.Listener.Close()
}
