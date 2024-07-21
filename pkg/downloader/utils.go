// sftp-server-test/utils.go

package downloader

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
