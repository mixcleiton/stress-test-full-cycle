package stresstest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type configTeste struct {
	ctx            context.Context
	inicio         <-chan struct{}
	limite         chan<- struct{}
	resultado      chan<- resultado
	idConcorrencia int
}

type resultado struct {
	requisicoes int
	successes   int
	failures    map[string]int
}

type Config struct {
	mu           sync.Mutex
	Url          string
	Requisicoes  int
	Concorrencia int
	rodando      bool
}

func ExecutarTestes(config Config) {
	resultados := make(chan resultado, config.Concorrencia)
	start := make(chan struct{})
	limitChan := make(chan struct{}, config.Requisicoes)
	ctx, cancel := context.WithCancel(context.Background())

	go config.limit(ctx, cancel, limitChan)

	fmt.Println("Preparando concorrência")
	wg := &sync.WaitGroup{}
	for i := 0; i < config.Concorrencia; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			configTeste := configTeste{
				ctx:            ctx,
				inicio:         start,
				limite:         limitChan,
				resultado:      resultados,
				idConcorrencia: i,
			}

			config.realizarTeste(configTeste)
		}(i)
	}

	fmt.Println("Rodando os testes")
	startTime := time.Now()
	close(start)
	wg.Wait()
	endTime := time.Since(startTime)

	finalResult := resultado{
		failures: map[string]int{},
	}
	for i := 0; i < config.Concorrencia; i++ {
		res := <-resultados
		finalResult.successes += res.successes
		finalResult.requisicoes += res.requisicoes

		for code, count := range res.failures {
			_, ok := finalResult.failures[code]
			if !ok {
				finalResult.failures[code] = count
			} else {
				finalResult.failures[code] = finalResult.failures[code] + count
			}
		}
	}

	fmt.Printf("Tempo total gasto na execução: %s\n", endTime.String())
	fmt.Printf("Quantidade total de requests realizados: %d\n", finalResult.requisicoes)
	fmt.Printf("Quantidade de requests com status HTTP 200: %d\n", finalResult.successes)
	fmt.Println("Distribuição de outros códigos de status HTTP:")
	for status, count := range finalResult.failures {
		fmt.Printf("HTTP %s: %d\n", status, count)
	}
}

func (cfg *Config) realizarTeste(configTeste configTeste) {
	resultados := resultado{
		failures: map[string]int{},
	}
	<-configTeste.inicio

	for {
		select {
		case <-configTeste.ctx.Done():
			configTeste.resultado <- resultados
			return
		default:
			client := http.DefaultClient
			req, _ := http.NewRequestWithContext(configTeste.ctx, "GET", cfg.Url, nil)
			response, err := client.Do(req)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					configTeste.limite <- struct{}{}
					continue
				}
				fmt.Println("Error while making request: ", err)
				resultados.requisicoes++
				resultados.failures["unknown error"]++
				configTeste.limite <- struct{}{}
				continue
			}

			if response.StatusCode == http.StatusOK {
				resultados.successes++
			} else {
				_, ok := resultados.failures[response.Status]
				if ok {
					resultados.failures[response.Status]++
				} else {
					resultados.failures[response.Status] = 1
				}
			}
			resultados.requisicoes++
			configTeste.limite <- struct{}{}
		}
	}
}

func (cfg *Config) limit(ctx context.Context, cancel context.CancelFunc, limitChan <-chan struct{}) {
	var total int
	for {
		select {
		case <-ctx.Done():
			return
		case <-limitChan:
			total++
			if total >= cfg.Requisicoes {
				cancel()
				return
			}
		}
	}
}
