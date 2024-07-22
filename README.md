Aplicação para Teste de Stress

Como utilizar:
É necessário primeiro fazer o build a aplicação, usando:
docker build -t server .

Depois de finalizado o build da imagem, podemos utilizar a aplicação utilizando o comando:
docker run server --url=http://google.com --requests=1 --concurrency=1

A váriavel url é obrigatória, e caso não seja informada requests e concurrency, será utilizado o valor 1 para elas.

No final a aplicação irá printar as seguintes informações:
Tempo total gasto na execução:
Quantidade total de requests realizados:
Quantidade de requests com status HTTP 200:
Distribuição de outros códigos de status HTTP: