package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// --- DECLARES GLOBAIS PARA SINCRONIZAÇÃO ---
var (
	produtosGlobais   []Produto
	historicoVendas   []Venda
	tabelaVisivel     *widget.Table
	produtosFiltrados []Produto // Controla o que aparece na busca
	textoBusca        string
	janelaPrincipal   fyne.Window
)

// --- DEFINIÇÕES DE TIPOS ---
type Medida int

const (
	Unidade Medida = iota
	Quilo
)

func (u Medida) String() string {
	if u == Quilo {
		return "kg"
	}
	return "un"
}

type Status int

const (
	Disponivel Status = iota
	Indisponivel
)

func (s Status) String() string {
	if s == Disponivel {
		return "Disponível"
	}
	return "Indisponível"
}

type Categoria int

const (
	Acougue Categoria = iota
	Hortifruti
	Limpeza
	Bebidas
	Mercearia
	Padaria
	Laticinios
)

func (c Categoria) String() string {
	return [...]string{"Açougue", "Hortifruti", "Limpeza", "Bebidas", "Mercearia", "Padaria", "Laticínios"}[c]
}

type Produto struct {
	Name       string    `json:"nome"`
	Cat        Categoria `json:"categoria"`
	Price      float64   `json:"preco"`
	Quantidade float64   `json:"quantidade"`
	TipoMedida Medida    `json:"medida"`
	Stock      Status    `json:"status"`
	CreatedAt  time.Time `json:"data_cadastro"`
}

func (p Produto) TotalItem() float64 { return p.Price * p.Quantidade }

func (p Produto) String() string {
	dataTexto := p.CreatedAt.Format("02/01/2006")
	totalItem := p.Price * p.Quantidade
	return fmt.Sprintf("Nome: %-12s | Cat: %-10s | Preço: R$ %6.2f | Qtd: %g%-2s | Total: R$ %7.2f | Status: %-12s | Data: %s", p.Name, p.Cat.String(), p.Price, p.Quantidade, p.TipoMedida, totalItem, p.Stock.String(), dataTexto)
}

// Venda registra o histórico de saídas do estoque
type Venda struct {
	Name       string    `json:"nome"`
	Cat        Categoria `json:"categoria"`
	Price      float64   `json:"preco_venda"` // Preço praticado no momento da venda
	Quantidade float64   `json:"quantidade"`
	TipoMedida Medida    `json:"medida"`
	SoldAt     time.Time `json:"data_venda"`
}

func main() {
	meuApp := app.New()
	janelaPrincipal = meuApp.NewWindow("Controle de Estoque Inteligente")
	janelaPrincipal.Resize(fyne.NewSize(900, 800))

	var errp, errv error
	produtosGlobais, errp = carregarProdutos()
	historicoVendas, errv = carregarVendasJSON()

	if errp != nil || errv != nil {
		produtosGlobais = []Produto{}
		historicoVendas = []Venda{}
	}

	produtosFiltrados = produtosGlobais

	telaLista := criarTelaListagem()
	telaOperacoes := criarTelaOperacoes()
	telaRelatorio := criarTelaRelatorio()
	telaVendas := criarTelaVendas()

	abas := container.NewAppTabs(
		container.NewTabItem("Listar Produtos", telaLista),
		container.NewTabItem("Operações Estoque", telaOperacoes),
		container.NewTabItem("Relatório TXT", telaRelatorio),
		container.NewTabItem("Relatório de Vendas", telaVendas),
	)
	abas.SetTabLocation(container.TabLocationTop)

	janelaPrincipal.SetContent(abas)
	janelaPrincipal.ShowAndRun()
}

func atualizarInterface() {
	if textoBusca == "" {
		produtosFiltrados = produtosGlobais
	} else {
		produtosFiltrados = []Produto{}
		for _, p := range produtosGlobais {
			if strings.Contains(strings.ToLower(p.Name), strings.ToLower(textoBusca)) {
				produtosFiltrados = append(produtosFiltrados, p)
			}
		}
	}
	if tabelaVisivel != nil {
		tabelaVisivel.Refresh()
	}
}

// --- CONSTRUTORES DE TELAS ---

func criarTelaListagem() fyne.CanvasObject {
	cabecalhos := []string{"ID", "Nome", "Categoria", "Preço", "Qtd", "Total", "Status", "Data Cadastro"}

	inputBusca := widget.NewEntry()
	inputBusca.SetPlaceHolder("Pesquisar produto por nome...")
	inputBusca.OnChanged = func(texto string) {
		textoBusca = texto
		atualizarInterface()
	}

	tabelaVisivel = widget.NewTable(
		func() (int, int) { return len(produtosFiltrados) + 1, len(cabecalhos) },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("Template")
			lbl.Alignment = fyne.TextAlignLeading
			return lbl
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)

			if id.Row == 0 {
				label.SetText(cabecalhos[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				return
			}

			p := produtosFiltrados[id.Row-1]
			label.TextStyle = fyne.TextStyle{Bold: false}

			idReal := -1
			for idx, orig := range produtosGlobais {
				if orig.Name == p.Name && orig.CreatedAt.Equal(p.CreatedAt) {
					idReal = idx
					break
				}
			}

			switch id.Col {
			case 0:
				label.SetText(fmt.Sprintf("%d", idReal))
			case 1:
				label.SetText(p.Name)
			case 2:
				label.SetText(p.Cat.String())
			case 3:
				label.SetText(fmt.Sprintf("R$ %.2f", p.Price))
			case 4:
				label.SetText(fmt.Sprintf("%g %s", p.Quantidade, p.TipoMedida.String()))
			case 5:
				label.SetText(fmt.Sprintf("R$ %.2f", p.TotalItem()))
			case 6:
				label.SetText(p.Stock.String())
			case 7:
				label.SetText(p.CreatedAt.Format("02/01/2006"))
			}
		},
	)

	tabelaVisivel.SetColumnWidth(0, 80)
	tabelaVisivel.SetColumnWidth(1, 140)
	tabelaVisivel.SetColumnWidth(2, 100)
	tabelaVisivel.SetColumnWidth(3, 90)
	tabelaVisivel.SetColumnWidth(4, 80)
	tabelaVisivel.SetColumnWidth(5, 100)
	tabelaVisivel.SetColumnWidth(6, 100)
	tabelaVisivel.SetColumnWidth(7, 110)

	topo := container.NewVBox(
		widget.NewLabelWithStyle("Estoque Atual", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		inputBusca,
	)

	return container.NewBorder(topo, nil, nil, nil, container.NewScroll(tabelaVisivel))
}

func criarTelaOperacoes() fyne.CanvasObject {
	inputNome := widget.NewEntry()
	inputNome.SetPlaceHolder("Nome do Produto")

	inputPreco := widget.NewEntry()
	inputPreco.SetPlaceHolder("Preço (Ex: 5.50)")

	inputQtd := widget.NewEntry()
	inputQtd.SetPlaceHolder("Quantidade")

	categoriasTexto := []string{"Açougue", "Hortifruti", "Limpeza", "Bebidas", "Mercearia", "Padaria", "Laticínios"}
	selectCat := widget.NewSelect(categoriasTexto, nil)
	selectCat.SetSelectedIndex(0)

	medidasTexto := []string{"un", "kg"}
	selectMedida := widget.NewSelect(medidasTexto, nil)
	selectMedida.SetSelectedIndex(0)

	form := widget.NewForm(
		widget.NewFormItem("Nome do Produto", inputNome),
		widget.NewFormItem("Categoria", selectCat),
		widget.NewFormItem("Preço (R$)", inputPreco),
		widget.NewFormItem("Quantidade", inputQtd),
		widget.NewFormItem("Unidade de Medida", selectMedida),
	)

	btnAdicionar := widget.NewButton("Adicionar Produto", func() {
		// Remove espaços em branco acidentais antes de validar
		nomeLimpo := strings.TrimSpace(inputNome.Text)
		precoTexto := strings.TrimSpace(inputPreco.Text)
		qtdTexto := strings.TrimSpace(inputQtd.Text)

		// 1. TRATAMENTO DE ERRO: Garante que nenhum campo foi enviado em branco
		if nomeLimpo == "" || precoTexto == "" || qtdTexto == "" {
			dialog.ShowError(fmt.Errorf("Por favor, preencha todos os campos!"), janelaPrincipal)
			return
		}

		// Corrige a digitação de vírgulas para o padrão de ponto flutuante do Go
		precoTexto = strings.ReplaceAll(precoTexto, ",", ".")
		qtdTexto = strings.ReplaceAll(qtdTexto, ",", ".")

		// 2. TRATAMENTO DE ERRO: Valida o formato do preço (rejeita letras ou fórmulas malucas)
		preco, err := strconv.ParseFloat(precoTexto, 64)
		if err != nil || preco < 0 {
			dialog.ShowError(fmt.Errorf("Preço inválido! Digite apenas números positivos (Ex: 5.50)."), janelaPrincipal)
			return
		}

		// 3. TRATAMENTO DE ERRO: Valida o formato da quantidade
		qtd, err := strconv.ParseFloat(qtdTexto, 64)
		if err != nil || qtd < 0 {
			dialog.ShowError(fmt.Errorf("Quantidade inválida! Digite apenas números positivos (Ex: 10 ou 2.5)."), janelaPrincipal)
			return
		}

		p := Produto{
			Name:       inputNome.Text,
			Cat:        Categoria(selectCat.SelectedIndex()),
			Price:      preco,
			Quantidade: qtd,
			TipoMedida: Medida(selectMedida.SelectedIndex()),
			CreatedAt:  time.Now(),
		}

		if qtd > 0 {
			p.Stock = Disponivel
		} else {
			p.Stock = Indisponivel
		}

		produtosGlobais = append(produtosGlobais, p)
		salvarJSON(produtosGlobais)
		atualizarInterface()

		inputNome.SetText("")
		inputPreco.SetText("")
		inputQtd.SetText("")
		dialog.ShowInformation("Sucesso", "Produto adicionado ao estoque!", janelaPrincipal)
	})

	btnRemover := widget.NewButton("Remover por ID", func() {
		dialog.ShowEntryDialog("Remover Produto", "Digite o número do ID Original do produto:", func(texto string) {
			// Remove espaços em branco acidentais antes e depois do texto
			texto = strings.TrimSpace(texto)

			// TRATAMENTO DE ERRO: Verifica se o usuário digitou letras, caracteres ou fórmulas inválidas
			idx, err := strconv.Atoi(texto)
			if err != nil {
				dialog.ShowError(fmt.Errorf("O ID deve conter apenas números inteiros!"), janelaPrincipal)
				return
			}

			// TRATAMENTO DE ERRO: Garante que o índice existe dentro dos limites do array
			if idx < 0 || idx >= len(produtosGlobais) {
				dialog.ShowError(fmt.Errorf("ID não encontrado no sistema! Use o ID Original listado na tabela."), janelaPrincipal)
				return
			}

			// Executa a remoção com segurança
			produtosGlobais = append(produtosGlobais[:idx], produtosGlobais[idx+1:]...)
			salvarJSON(produtosGlobais)
			atualizarInterface()
			dialog.ShowInformation("Sucesso", "Produto removido com sucesso!", janelaPrincipal)
		}, janelaPrincipal)
	})

	btnEstoque := widget.NewButton("Movimentar Quantidade", func() {
		dialog.ShowEntryDialog("Ajustar Estoque", "Digite o ID Original do produto:", func(idxTexto string) {
			// Remove espaços em branco acidentais
			idxTexto = strings.TrimSpace(idxTexto)

			// TRATAMENTO DE ERRO: Verifica se o usuário digitou letras no ID
			idx, err := strconv.Atoi(idxTexto)
			if err != nil {
				dialog.ShowError(fmt.Errorf("O ID deve conter apenas números inteiros!"), janelaPrincipal)
				return
			}

			if idx < 0 || idx >= len(produtosGlobais) {
				dialog.ShowError(fmt.Errorf("ID não encontrado no sistema!"), janelaPrincipal)
				return
			}

			dialog.ShowEntryDialog("Quantidade", "Quantidade (+ entrada, - saída):", func(qtdTexto string) {
				// Remove espaços em branco acidentais
				qtdTexto = strings.TrimSpace(qtdTexto)

				// TRATAMENTO DE ERRO: Impede entradas malucas como "0004-4" ou caracteres inválidos
				// Substitui vírgulas por pontos caso o usuário digite "2,5" em vez de "2.5"
				qtdTexto = strings.ReplaceAll(qtdTexto, ",", ".")

				val, err := strconv.ParseFloat(qtdTexto, 64)
				if err != nil {
					dialog.ShowError(fmt.Errorf("Quantidade inválida! Digite apenas números (Ex: 5 ou -3.5)."), janelaPrincipal)
					return
				}

				// Impede que o usuário digite o número 0, pois não altera o estoque
				if val == 0 {
					dialog.ShowError(fmt.Errorf("A quantidade não pode ser zero!"), janelaPrincipal)
					return
				}

				// --- LÓGICA DE DETECÇÃO DE VENDA ---
				if val < 0 {
					qtdVendida := -val

					// TRATAMENTO DE ERRO: Impede vender mais do que o estoque atual possui
					if qtdVendida > produtosGlobais[idx].Quantidade {
						dialog.ShowError(fmt.Errorf("Estoque insuficiente! Estoque atual: %g", produtosGlobais[idx].Quantidade), janelaPrincipal)
						return
					}

					novaVenda := Venda{
						Name:       produtosGlobais[idx].Name,
						Cat:        produtosGlobais[idx].Cat,
						Price:      produtosGlobais[idx].Price,
						Quantidade: qtdVendida,
						TipoMedida: produtosGlobais[idx].TipoMedida,
						SoldAt:     time.Now(),
					}

					historicoVendas = append(historicoVendas, novaVenda)
					salvarVendasJSON(historicoVendas)
				}
				// ----------------------------------------

				produtosGlobais[idx].Quantidade += val
				if produtosGlobais[idx].Quantidade <= 0 {
					produtosGlobais[idx].Quantidade = 0
					produtosGlobais[idx].Stock = Indisponivel
				} else {
					produtosGlobais[idx].Stock = Disponivel
				}
				salvarJSON(produtosGlobais)
				atualizarInterface()
				dialog.ShowInformation("Sucesso", "Movimentação executada com sucesso!", janelaPrincipal)
			}, janelaPrincipal)
		}, janelaPrincipal)
	})

	btnPreco := widget.NewButton("Atualizar Preço", func() {
		dialog.ShowEntryDialog("Alterar Preço", "Digite o ID Original do produto:", func(idxTexto string) {
			// Remove espaços em branco acidentais
			idxTexto = strings.TrimSpace(idxTexto)

			// TRATAMENTO DE ERRO: Verifica se o usuário digitou letras ou caracteres inválidos no ID
			idx, err := strconv.Atoi(idxTexto)
			if err != nil {
				dialog.ShowError(fmt.Errorf("O ID deve conter apenas números inteiros!"), janelaPrincipal)
				return
			}

			if idx < 0 || idx >= len(produtosGlobais) {
				dialog.ShowError(fmt.Errorf("ID não encontrado no sistema!"), janelaPrincipal)
				return
			}

			dialog.ShowEntryDialog("Novo Preço", "Digite o novo preço (R$):", func(precoTexto string) {
				// Remove espaços em branco acidentais
				precoTexto = strings.TrimSpace(precoTexto)

				// TRATAMENTO DE ERRO: Substitui vírgulas por pontos caso digitem "5,50"
				precoTexto = strings.ReplaceAll(precoTexto, ",", ".")

				nPreco, err := strconv.ParseFloat(precoTexto, 64)
				// TRATAMENTO DE ERRO: Impede letras, fórmulas inválidas ou valores negativos
				if err != nil || nPreco < 0 {
					dialog.ShowError(fmt.Errorf("Preço inválido! Digite apenas números positivos (Ex: 10.50)."), janelaPrincipal)
					return
				}

				produtosGlobais[idx].Price = nPreco
				salvarJSON(produtosGlobais)
				atualizarInterface()
				dialog.ShowInformation("Sucesso", "Preço atualizado com sucesso!", janelaPrincipal)
			}, janelaPrincipal)
		}, janelaPrincipal)
	})

	// CORREÇÃO: Adicionado o retorno do layout visual que estava ausente
	layoutBotoes := container.NewVBox(btnAdicionar, btnRemover, btnEstoque, btnPreco)
	return container.NewScroll(container.NewVBox(widget.NewLabel("Gerenciar Itens e Estoque"), form, layoutBotoes))
}

func criarTelaRelatorio() fyne.CanvasObject {
	textoRelatorio := widget.NewMultiLineEntry()
	textoRelatorio.SetPlaceHolder("Nenhum relatório gerado ainda...")
	textoRelatorio.Wrapping = fyne.TextWrapWord

	opcoesPeriodo := []string{"Todo o Estoque", "Diário", "Semanal", "Mensal"}
	selectPeriodo := widget.NewSelect(opcoesPeriodo, nil)
	selectPeriodo.SetSelectedIndex(0)

	formRelatorio := widget.NewForm(
		widget.NewFormItem("Período do Relatório:", selectPeriodo),
	)

	btnCarregar := widget.NewButton("Atualizar/Ler Relatório TXT", func() {
		periodoEscolhido := selectPeriodo.Selected
		produtosFiltradosParaRelatorio := filtraProdutosPorPeriodo(produtosGlobais, periodoEscolhido)

		if len(produtosFiltradosParaRelatorio) == 0 {
			dialog.ShowInformation("Aviso", fmt.Sprintf("Nenhuma movimentação ou produto encontrado para o período %s.", periodoEscolhido), janelaPrincipal)
			return
		}

		err := gerarRelatorioTxt(produtosFiltradosParaRelatorio, periodoEscolhido)
		if err != nil {
			dialog.ShowError(err, janelaPrincipal)
			return
		}

		caminhoTxt := obterCaminhoRelatorio()
		dados, err := ioutil.ReadFile(caminhoTxt)
		if err != nil {
			textoRelatorio.SetText("Erro ao abrir arquivo. Certifique-se de que clicou em 'Gerar Relatório IA' na aba de Operações.")
			return
		}
		textoRelatorio.SetText(string(dados))

		dialog.ShowInformation("Sucesso", fmt.Sprintf("Relatório %s exportado com sucesso!", periodoEscolhido), janelaPrincipal)
	})

	return container.NewBorder(
		container.NewVBox(widget.NewLabel("Visualizador de Relatório"), formRelatorio, btnCarregar), nil, nil, nil,
		textoRelatorio)
}

func criarTelaVendas() fyne.CanvasObject {
	textoVendas := widget.NewMultiLineEntry()
	textoVendas.SetPlaceHolder("Nenhum relatório de vendas gerado para o período selecionado...")
	textoVendas.Wrapping = fyne.TextWrapWord

	opcoesPeriodo := []string{"Todo o Histórico", "Diário", "Semanal", "Mensal"}
	selectPeriodoVendas := widget.NewSelect(opcoesPeriodo, nil)
	selectPeriodoVendas.SetSelectedIndex(0)

	btnAnalisar := widget.NewButton("Analisar Faturamento e Vendas", func() {
		// 1. Filtra as vendas baseado no período escolhido
		var vendasFiltradas []Venda
		agora := time.Now()
		periodo := selectPeriodoVendas.Selected

		for _, v := range historicoVendas {
			switch periodo {
			case "Diário":
				if v.SoldAt.Year() == agora.Year() && v.SoldAt.YearDay() == agora.YearDay() {
					vendasFiltradas = append(vendasFiltradas, v)
				}
			case "Semanal":
				if agora.Sub(v.SoldAt) <= 7*24*time.Hour {
					vendasFiltradas = append(vendasFiltradas, v)
				}
			case "Mensal":
				if v.SoldAt.Year() == agora.Year() && v.SoldAt.Month() == agora.Month() {
					vendasFiltradas = append(vendasFiltradas, v)
				}
			default: // "Todo o Histórico"
				vendasFiltradas = append(vendasFiltradas, v)
			}
		}

		if len(vendasFiltradas) == 0 {
			textoVendas.SetText(fmt.Sprintf("Nenhuma venda registrada no período: %s", periodo))
			return
		}

		// 2. Processa os totais e monta a string do relatório
		var construtor strings.Builder
		construtor.WriteString(fmt.Sprintf("=================== ANÁLISE DE VENDAS (%s) ===================\n\n", strings.ToUpper(periodo)))

		resumoFaturamento := make(map[Categoria]float64)
		resumoQuantidades := make(map[string]float64)
		faturamentoTotal := 0.0

		for _, v := range vendasFiltradas {
			dataTexto := v.SoldAt.Format("02/01/2006 15:04")
			totalVendaItem := v.Price * v.Quantidade

			construtor.WriteString(fmt.Sprintf("Item: %-12s | Qtd: %g%-2s | Valor Un: R$ %6.2f | Total: R$ %7.2f | Data: %s\n",
				v.Name, v.Quantidade, v.TipoMedida.String(), v.Price, totalVendaItem, dataTexto))

			resumoFaturamento[v.Cat] += totalVendaItem
			faturamentoTotal += totalVendaItem
			resumoQuantidades[v.Name] += v.Quantidade
		}

		construtor.WriteString("\n=================== PRODUTOS MAIS VENDIDOS ===================\n\n")
		for nome, qtd := range resumoQuantidades {
			construtor.WriteString(fmt.Sprintf("Produto: %-15s | Total Vendido: %g\n", nome, qtd))
		}

		construtor.WriteString("\n=================== FATURAMENTO POR CATEGORIA ===================\n\n")
		for i := 0; i <= 6; i++ {
			c := Categoria(i)
			construtor.WriteString(fmt.Sprintf("%-12s: R$ %10.2f Faturados\n", c.String(), resumoFaturamento[c]))
		}

		construtor.WriteString("\n----------------------------------------------------\n")
		construtor.WriteString(fmt.Sprintf("FATURAMENTO TOTAL DO PERÍODO: R$ %10.2f\n", faturamentoTotal))

		// 3. Exibe o resultado direto na tela para o usuário
		relatorioTextoCompleto := construtor.String()
		textoVendas.SetText(relatorioTextoCompleto)

		// --- NOVA LÓGICA: CRIA E SALVA O ARQUIVO TXT AUTOMATICAMENTE ---
		caminhoEstoque := obterCaminhoRelatorio()
		caminhoVendas := filepath.Join(filepath.Dir(caminhoEstoque), "relatorio_vendas.txt")

		err := os.WriteFile(caminhoVendas, []byte(relatorioTextoCompleto), 0644)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Erro ao salvar o arquivo TXT: %v", err), janelaPrincipal)
			return
		}

		// Exibe um aviso de sucesso discreto na janela para o usuário saber onde o arquivo foi salvo
		dialog.ShowInformation("Sucesso", fmt.Sprintf("Relatório de vendas salvo em:\n%s", caminhoVendas), janelaPrincipal)
	})

	topo := container.NewVBox(
		widget.NewLabelWithStyle("Histórico de Faturamento", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewForm(widget.NewFormItem("Filtrar Período:", selectPeriodoVendas)),
		btnAnalisar,
	)

	return container.NewBorder(topo, nil, nil, nil, textoVendas)
}

func carregarProdutos() ([]Produto, error) {
	produtos, err := carregarJSON()
	if err != nil {
		return nil, err
	}
	return produtos, nil
}

func obterCaminhoBanco() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "estoque.json"
	}
	appDir := filepath.Join(configDir, "meu-estoque")
	os.MkdirAll(appDir, 0755)
	return filepath.Join(appDir, "estoque.json")
}

func obterPastaDocuments() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	pastaDocumentos := filepath.Join(homeDir, "Documents")
	return pastaDocumentos
}

func obterCaminhoRelatorio() string {
	pastaDocumentos := obterPastaDocuments()

	os.MkdirAll(pastaDocumentos, 0755)
	return filepath.Join(pastaDocumentos, "relatorio.txt")
}

func salvarJSON(produtos []Produto) error {
	dados, _ := json.MarshalIndent(produtos, "", "  ")
	return os.WriteFile(obterCaminhoBanco(), dados, 0644)
}

func carregarJSON() ([]Produto, error) {
	dados, err := os.ReadFile(obterCaminhoBanco())
	if err != nil {
		return nil, err
	}
	var produtos []Produto
	json.Unmarshal(dados, &produtos)
	return produtos, nil
}

func salvarVendasJSON(vendas []Venda) error {
	dados, _ := json.MarshalIndent(vendas, "", "  ")

	// Utiliza a mesma lógica da sua função obterCaminhoBanco, mas direciona para o arquivo vendas.json
	configDir, err := os.UserConfigDir()
	if err != nil {
		return os.WriteFile("vendas.json", dados, 0644)
	}
	caminhoVendas := filepath.Join(configDir, "meu-estoque", "vendas.json")

	return os.WriteFile(caminhoVendas, dados, 0644)
}

func carregarVendasJSON() ([]Venda, error) {
	configDir, err := os.UserConfigDir()
	var caminhoVendas string
	if err != nil {
		caminhoVendas = "vendas.json"
	} else {
		caminhoVendas = filepath.Join(configDir, "meu-estoque", "vendas.json")
	}

	dados, err := os.ReadFile(caminhoVendas)
	if err != nil {
		// Se o arquivo ainda não existir (primeira venda), retorna uma lista vazia sem dar erro
		return []Venda{}, nil
	}

	var vendas []Venda
	json.Unmarshal(dados, &vendas)
	return vendas, nil
}

// CORREÇÃO: Alterado de 'produtos' para 'produtosFiltrados' dentro do loop
func gerarRelatorioTxt(produtosFiltrados []Produto, periodo string) error {
	arquivo, err := os.Create(obterCaminhoRelatorio())
	if err != nil {
		return err
	}
	defer arquivo.Close()

	arquivo.WriteString(fmt.Sprintf("========================= RELATÓRIO DETALHADO DE ESTOQUE (%s) ==========\n\n", strings.ToUpper(periodo)))
	resumoMap := make(map[Categoria]float64)
	totalGeral := 0.0

	for _, p := range produtosFiltrados {
		arquivo.WriteString(p.String() + "\n")
		valorNoEstoque := p.Price * p.Quantidade
		resumoMap[p.Cat] += valorNoEstoque
		totalGeral += valorNoEstoque
	}

	arquivo.WriteString("\n========================= RESUMO FINANCEIRO POR CATEGORIA ==========\n\n")
	for i := 0; i <= 6; i++ {
		c := Categoria(i)
		arquivo.WriteString(fmt.Sprintf("%-12s: R$ %10.2f\n", c.String(), resumoMap[c]))
	}
	arquivo.WriteString("\n----------------------------------------------------\n")
	arquivo.WriteString(fmt.Sprintf("VALOR TOTAL EM ESTOQUE: R$ %10.2f\n", totalGeral))
	return nil
}

func filtraProdutosPorPeriodo(produtos []Produto, periodo string) []Produto {
	var filtrados []Produto
	agora := time.Now()

	for _, p := range produtos {
		switch periodo {
		case "Diário":
			if p.CreatedAt.Year() == agora.Year() && p.CreatedAt.YearDay() == agora.YearDay() {
				filtrados = append(filtrados, p)
			}
		case "Semanal":
			if agora.Sub(p.CreatedAt) <= 7*24*time.Hour {
				filtrados = append(filtrados, p)
			}
		case "Mensal":
			if p.CreatedAt.Year() == agora.Year() && p.CreatedAt.Month() == agora.Month() {
				filtrados = append(filtrados, p)
			}
		default:
			filtrados = append(filtrados, p)
		}
	}
	return filtrados
}
