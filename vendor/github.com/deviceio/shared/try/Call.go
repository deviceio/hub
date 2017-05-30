package try

// Call invokes tryf for in which case if an error is returned or a panic is
// encountered the catchf function is invoked.
func Call(tryf func() error, catchf func(error, string)) {
	(&tryCatch{
		tryf:   tryf,
		catchf: catchf,
	}).run()
}
