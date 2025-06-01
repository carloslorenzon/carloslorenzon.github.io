---
layout: post
title:  "Por que você deveria usar Circuit Breaker nos seus projetos?"
categories: blog hard-skills resiliencia
author: Carlos Lorenzon
comments: true
---
Se você já trabalhou com sistemas modernos, especialmente aqueles que usam microsserviços ou dependem de APIs externas, sabe que garantir que tudo funcione bem o tempo todo é um baita desafio. 

E um dos maiores vilões nessas situações são as falhas inesperadas que podem derrubar todo o sistema. É aí que entra o Circuit Breaker, ou “disjuntor”, um padrão de design que pode salvar o seu projeto na hora do aperto.

## O que é Circuit Breaker, afinal?

Pense no disjuntor da sua casa: quando algo dá problema na rede elétrica, ele desarma para proteger todo o sistema contra sobrecarga. 

O Circuit Breaker faz algo parecido para o seu software. Ele monitora as chamadas que sua aplicação faz para serviços externos e, quando percebe que algo está errado, por exemplo: um serviço está fora do ar ou lento demais, ele "desliga" essas tentativas, impedindo que sua aplicação fique presa esperando e piorando a situação.

## Quais os benefícios de usar esse padrão?

**1 - Evita o famoso efeito dominó**

Sem o Circuit Breaker, uma falha em um serviço pode se espalhar, derrubando tudo ao redor. Com ele, o problema é isolado logo no começo.

**2 - Melhora a experiência do usuário**

Ninguém gosta de sistema travando ou carregando eternamente, certo? O Circuit Breaker permite respostas rápidas, avisando que o serviço está indisponível ou usando dados em cache, por exemplo.

**3 - Economiza recursos do sistema**

Ficar insistindo em chamadas que falham só gasta CPU, memória e conexões desnecessariamente. O disjuntor evita isso.

**4 - Permite uma recuperação segura**

O Circuit Breaker não fica "desligado" para sempre. Ele testa periodicamente se o serviço voltou ao normal, reativando as chamadas gradualmente para garantir que tudo esteja estável.

**5 - Facilita o monitoramento**

Ferramentas que implementam Circuit Breaker normalmente já vêm com métricas e dashboards para você acompanhar o comportamento e agir rápido quando algo acontecer.

### Exemplos práticos

Imagine um e-commerce: se o serviço de pagamento ficar fora do ar, ao invés de travar o site inteiro, o Circuit Breaker pode desativar temporariamente essa opção e avisar o usuário, deixando ele continuar navegando.

Ou pense numa API que depende de um microserviço para recomendações. Se esse serviço ficar lento, o disjuntor evita que essa lentidão atrapalhe toda a resposta da API.

### Como colocar isso pra funcionar?

Depende da linguagem e da arquitetura, mas aqui vão algumas opções populares:

- Java: Resilience4j (o Hystrix está meio abandonado)
- .NET: Polly
- Node.js: opossum
- Go: sony/gobreaker, goresilience

Exemplo de implementação usando Go

```go
package main

import (
    "errors"
    "fmt"
    "math/rand"
    "time"

    "github.com/sony/gobreaker"
)

// Simula uma chamada a um serviço externo que pode falhar aleatoriamente
func unreliableService() (string, error) {
    if rand.Float32() < 0.7 { // 70% de chance de falhar
        return "", errors.New("serviço indisponível")
    }
    return "dados recebidos com sucesso", nil
}

func main() {
    rand.Seed(time.Now().UnixNano())

    // Configurando o Circuit Breaker
    cbSettings := gobreaker.Settings{
        Name:        "MeuCircuitBreaker",
        MaxRequests: 3, // Quantidade de chamadas permitidas no estado half-open
        Interval:    0, // Sem limpeza automática do contador de falhas
        Timeout:     5 * time.Second, // Tempo para tentar fechar o circuito após aberto
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            // Abre o circuito após 3 falhas consecutivas
            return counts.ConsecutiveFailures >= 3
        },
        OnStateChange: func(name string, from, to gobreaker.State) {
            fmt.Printf("Circuit Breaker mudou de %s para %s\n", from.String(), to.String())
        },
    }

    cb := gobreaker.NewCircuitBreaker(cbSettings)

    // Simulação de várias chamadas ao serviço, usando o Circuit Breaker
    for i := 1; i <= 15; i++ {
        result, err := cb.Execute(func() (interface{}, error) {
            return unreliableService()
        })

        if err != nil {
            fmt.Printf("Chamada %d: falha - %s\n", i, err)
        } else {
            fmt.Printf("Chamada %d: sucesso - %s\n", i, result.(string))
        }

        time.Sleep(1 * time.Second)
    }
}
```

### Resumo: por que investir no Circuit Breaker?

No fim das contas, o Circuit Breaker não é um luxo, mas uma necessidade para quem quer construir sistemas realmente confiáveis e com boa performance. Ele não impede que falhas aconteçam, afinal ninguém tem controle total sobre tudo, mas ajuda você a lidar com essas falhas de maneira inteligente, evitando que um problema vire uma bola de neve.

Se você está criando APIs, microsserviços ou qualquer aplicação que dependa de terceiros, vale muito a pena conhecer e implementar esse padrão. Seu sistema (e seus usuários) agradecem!