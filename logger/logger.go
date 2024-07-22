// logger/logger.go
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// loggerInstance é uma variável global para armazenar a instância do logger.
// variáveis que começam com letra minúscula são privadas ao pacote.
var loggerInstance *zerolog.Logger

// InitLogger inicializa o logger global.
// Esse sim é constructor
func InitLogger() {
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	// Atribui o endereço da variável local 'l' à variável global 'loggerInstance'.
	// Isso faz com que 'loggerInstance' aponte para a instância do logger criada.
	// O '&' é usado para obter o endereço de memória da variável 'l'.
	// Em Go, isso é uma forma comum de compartilhar uma instância entre diferentes partes do código.
	loggerInstance = &l
}

// GetLogger retorna um ponteiro para a instância global do logger.
// Esta função atua como um getter, permitindo o acesso ao logger em outras partes do código.
// O retorno de um ponteiro (*zerolog.Logger) é comum em Go para evitar cópias desnecessárias de objetos grandes.
// Em Java, isso seria semelhante a um método estático que retorna uma instância compartilhada (padrão Singleton)
func GetLogger() *zerolog.Logger {
	return loggerInstance
}
