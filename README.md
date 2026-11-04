# Experimentos de Tempo em Go

## O que é

Este projeto testa se é possível fazer um programa em Go perceber o tempo passando mais rápido utilizando recursos nativos do Linux. O objetivo é verificar na prática se é possível acelerar a passagem do tempo para o programa através de dois métodos:

1. **Experimento 1:** Definindo um *offset* em um *time namespace*.
2. **Experimento 2:** Alterando o `CLOCK_REALTIME` da máquina.

**Nota:** Os resultados detalhados obtidos nos testes originais estão disponíveis nos arquivos de log deste repositório (`base.txt`, `namespacechange.txt` e `exp2.txt`). A seção [Resultados e Conclusões](#resultados-e-conclusões) traz o resumo do que foi observado.

## Pré-requisitos

* **Sistema:** Linux x86_64 com kernel >= 5.6 e *time namespaces* habilitados.
* **Ferramentas:** `unshare` (com suporte a `-T` / `--time`, `--monotonic` e `--boottime`), linguagem Go instalada e acesso ao `sudo`.
* **ATENÇÃO:** É recomendável usar **Máquina Virtual (VM) descartável** para o experimento 2. A execução alterará a hora global do sistema inteiro, o que afeta cron, logs, certificados e outros serviços.

## Como rodar (Para reprodução)

### 1. Compilar o código de teste

Como o arquivo `timetest.go` já está no repositório, basta compilá-lo como um binário estático:

```bash
CGO_ENABLED=0 go build -o timetest timetest.go
```

### 2. Rodar o teste normal (Linha de base)

Veja como o programa se comporta sem nenhuma alteração no tempo:

```bash
./timetest
```

### 3. Experimento 1 (Time Namespace)

#### 3.1. Iniciar o programa com offset

Inicie o programa aplicando um deslocamento (*offset*) nos relógios monotônico e de boot usando *namespaces*:

```bash
sudo unshare --time --fork --monotonic 100000 --boottime 100000 ./timetest
```

#### 3.2. Tentar alterar o offset com o processo em execução

Para verificar se é possível mudar o "salto" de tempo de um processo que já está rodando, execute:

```bash
sudo unshare -T -f --monotonic 100000 sleep 120 &
sleep 1
cat /proc/$(pgrep -n sleep)/timens_offsets
echo "monotonic 500000 0" | sudo tee /proc/$(pgrep -n sleep)/timens_offsets
```

### 4. Experimento 2 (Alterar relógio da máquina)

Neste teste, você desativa a sincronização de tempo padrão e força um avanço de 9 segundos no relógio do sistema a cada segundo real, rodando o programa em paralelo:

```bash
sudo timedatectl set-ntp false
./timetest &
sudo bash -c 'for i in $(seq 1 10); do sleep 1; date -s @$(( $(date +%s) + 9 )) >/dev/null; done'
wait
sudo timedatectl set-ntp true
```

---

## Resultados e Conclusões

Os logs completos de execução estão salvos nos arquivos `.txt` anexos ao repositório. Abaixo está o resumo do que os dados demonstraram:

* **Linha de Base:** O programa roda normalmente, e todas as marcações de tempo e *sleeps* avançam uniformemente de 1 em 1 segundo.
* **Experimento 1 (Time Namespace):** Ao rodar com um *offset* de +100.000 segundos, o relógio monotônico absoluto realmente salta para refletir o deslocamento. No entanto, **a velocidade do relógio permanece inalterada** (continua variando exatamente 1 segundo por iteração). Além disso, tentar alterar o *namespace* de um processo que já está em execução falha com erro de permissão (`os error 13: Permission denied`), indicando que o kernel bloqueia mudanças dinâmicas no *offset*.
* **Experimento 2 (Alterando CLOCK_REALTIME):** Ao forçar saltos sucessivos na hora da máquina, o relógio global (hora de parede) avança de 10 em 10 segundos, mas **o andamento do programa e os tempos de execução não são afetados**. O Go (e suas funções como `time.Sleep` e `time.Since`) guia-se exclusivamente pelo `CLOCK_MONOTONIC`, que ignora os saltos do relógio do sistema.

**Conclusão Final:**
Os testes comprovam que o Go é blindado contra alterações no horário global do sistema (`CLOCK_REALTIME`) para controle de durações e *timers*. Embora os *Time Namespaces* permitam alterar o ponto de referência do relógio monotônico (adicionando um offset na inicialização), eles não servem para "acelerar" ou "desacelerar" a passagem do tempo para o processo.
