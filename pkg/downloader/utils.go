// pkg/downloader/utils.go
package downloader

// min retorna o menor de dois inteiros.
// Esta é uma função auxiliar usada para limitar o número de arquivos baixados.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
