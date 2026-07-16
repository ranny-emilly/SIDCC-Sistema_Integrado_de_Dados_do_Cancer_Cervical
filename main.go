package main

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	_ "://github.com"
)



func init() {
	rand.Seed(time.Now().UnixNano())
}

type Paciente struct {
	ID               int    `json:"id"`
	CartaoSUS        string `json:"cartao_sus"`
	CPFPaciente      string `json:"cpf_paciente"`
	NomeCompleto     string `json:"nome_completo"`
	DataNascimento   string `json:"data_nascimento"`
	CEP              string `json:"cep"`
	DDD              string `json:"ddd"`
	Telefone         string `json:"telefone"`
	Fixo             string `json:"fixo"`
	EmailPaciente    string `json:"email_paciente"`
	Nacionalidade    string `json:"nacionalidade"`
	UF               string `json:"uf"`
	RacaCor          string `json:"raca_cor"`
	Escolaridade     string `json:"escolaridade"`
	NomeMae          string `json:"nome_mae"`
	NomeSocial       string `json:"nome_social"`
	Logradouro       string `json:"logradouro"`
	NumeroResidencia string `json:"numero_residencia"`
	Complemento      string `json:"complemento"`
	Setor            string `json:"setor"`
	Cidade           string `json:"cidade"`
	CodMunicipio     string `json:"cod_municipio"`
	PontoReferencia  string `json:"ponto_referencia"`
}

type ExameCitopatologico struct {
	PacienteID             int            `json:"paciente_id"`
	MotivoExame            string         `json:"motivo_exame"`
	PrimeiraVezExame       string         `json:"primeira_vez_exame"`
	UsaDIU                 string         `json:"usa_diu"`
	UsaAnticoncepcional    string         `json:"usa_anticoncepcional"`
	EstaGestante           string         `json:"esta_gestante"`
	UsaHormonio            string         `json:"usa_hormonio"`
	JaFezRadioterapia      string         `json:"ja_fez_radioterapia"`
	DataUltimaMenstruacao string         `json:"data_ultima_menstruacao"`
	EstaMenopausa          string         `json:"esta_menopausa"`
	TeveCorrimento         string         `json:"teve_corrimento"`
	TeveSangramentoAnormal string         `json:"teve_sangramento_anormal"`
}

type UBSInfoRequest struct {
	UF            string `json:"uf"`
	Protocolo     string `json:"protocolo"`
	Cnes          string `json:"cnes"`
	Unidade       string `json:"unidade"`
	MunicipiosUbs string `json:"municipios_ubs"`
	Prontuario    string `json:"prontuario"`
}

type FichaCitopatologicaPayload struct {
	UBSInfoData UBSInfoRequest      `json:"ubs_info_data"`
	ExameData   ExameCitopatologico `json:"exame_data"`
}

type SacForm struct {
	Email    string `json:"email"`
	Mensagem string `json:"mensagem"`
}

func limparMascara(str string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, str)
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	return emailRegex.MatchString(email)
}

func conectar() (*sql.DB, error) {
	// Se o Render fornecer uma URL externa de banco de dados, o Go usará ela automaticamente.
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "host=localhost port=5432 user=postgres password=postgres dbname=ProjetoIntegrador sslmode=disable"
	}
	return sql.Open("postgres", connStr)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With, Accept, Origin, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func inserirPacienteHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	var paciente Paciente
	err := json.NewDecoder(r.Body).Decode(&paciente)
	if err != nil {
		http.Error(w, "Erro ao decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	paciente.CPFPaciente = limparMascara(paciente.CPFPaciente)
	paciente.CartaoSUS = limparMascara(paciente.CartaoSUS)
	paciente.CEP = limparMascara(paciente.CEP)
	paciente.Telefone = limparMascara(paciente.Telefone)
	paciente.DDD = limparMascara(paciente.DDD)
	paciente.Fixo = limparMascara(paciente.Fixo)

	if len(paciente.DDD) > 2 {
		paciente.DDD = paciente.DDD[:2]
	}
	if len(paciente.Telefone) > 9 {
		paciente.Telefone = paciente.Telefone[:9]
	}
	if len(paciente.CEP) > 8 {
		paciente.CEP = paciente.CEP[:8]
	}
	if len(paciente.Fixo) > 20 {
		paciente.Fixo = paciente.Fixo[:20]
	}

	var dataNascimentoTime time.Time
	if paciente.DataNascimento != "" {
		dataNascimentoTime, err = time.Parse("2006-01-02", paciente.DataNascimento)
		if err != nil {
			http.Error(w, "Formato de data de nascimento inválido (esperado YYYY-MM-DD): "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Data de nascimento é obrigatória", http.StatusBadRequest)
		return
	}

	var codMunicipio sql.NullString
	if paciente.CodMunicipio != "" {
		codMunicipio = sql.NullString{String: paciente.CodMunicipio, Valid: true}
	} else {
		codMunicipio = sql.NullString{Valid: false}
	}

	_, err = db.Exec(`
		INSERT INTO paciente_infos
		(cartao_sus, cpf_paciente, nome_completo, data_nascimento, cep, ddd, telefone, fixo, email, nacionalidade, uf, raca_cor, escolaridade, nome_mae, nome_social, logradouro, numero_residencia, complemento, setor, cidade, cod_municipio, ponto_referencia)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
	`,
		paciente.CartaoSUS, paciente.CPFPaciente, paciente.NomeCompleto, dataNascimentoTime,
		paciente.CEP, paciente.DDD, paciente.Telefone, paciente.Fixo, paciente.EmailPaciente, paciente.Nacionalidade, paciente.UF, paciente.RacaCor, paciente.Escolaridade, paciente.NomeMae, paciente.NomeSocial, paciente.Logradouro, paciente.NumeroResidencia, paciente.Complemento, paciente.Setor, paciente.Cidade, codMunicipio, paciente.PontoReferencia,
	)
	if err != nil {
		http.Error(w, "Erro ao inserir no banco: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Paciente cadastrado com sucesso!"))
}

func listarPacientesAPI(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	rows, err := db.Query("SELECT id, nome_completo, cpf_paciente FROM paciente_infos")
	if err != nil {
		http.Error(w, "Erro ao buscar pacientes: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pacientes []Paciente
	for rows.Next() {
		var p Paciente
		err := rows.Scan(&p.ID, &p.NomeCompleto, &p.CPFPaciente)
		if err == nil {
			pacientes = append(pacientes, p)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pacientes)
}

// ADICIONADO: Função principal que gerencia o servidor e as rotas
func main() {
	db, err := conectar()
	if err != nil {
		log.Fatal("Falha ao abrir conexao com banco:", err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	// Endpoints da API REST
	mux.HandleFunc("/api/pacientes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			inserirPacienteHandler(w, r, db)
		} else if r.Method == http.MethodGet {
			listarPacientesAPI(w, r, db)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})

	// Servidor de arquivos estáticos (Carrega seus arquivos HTML, CSS e JS da pasta "Statics")
	fileServer := http.FileServer(http.Dir("./Statics"))
	mux.Handle("/", fileServer)

	// Captura dinamicamente a porta definida pelo Render ou usa a 8080 localmente
	portaServidor := os.Getenv("PORT")
	if portaServidor == "" {
		portaServidor = "8080"
	}

	handlerComCORS := corsMiddleware(mux)

	fmt.Printf("Servidor rodando na porta %s...\n", portaServidor)
	log.Fatal(http.ListenAndServe(":"+portaServidor, handlerComCORS))
}
