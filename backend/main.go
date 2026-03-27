package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// ================================================================
// ESTRUCTURAS ORIGINALES + EXTENSIONES
// ================================================================

// Block es la unidad básica de la cadena (ORIGINAL)
type Block struct {
	Hash      []byte       `json:"hash"`
	Data      []byte       `json:"data"`
	PrevHash  []byte       `json:"prevHash"`
	Nonce     int          `json:"nonce"`
	Validator string       `json:"validator"`
	Timestamp time.Time    `json:"timestamp"`      // NUEVO: timestamp del bloque
	Record    *AuditRecord `json:"record"`         // NUEVO: registro estructurado
	MLScore   float64      `json:"mlScore"`        // NUEVO: score del validador
}

// AuditRecord es el nuevo tipo de dato estructurado que se almacena
type AuditRecord struct {
	Type        string                 `json:"type"`        // "log", "audit", "event"
	UserID      string                 `json:"userId"`
	Action      string                 `json:"action"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// Blockchain (ORIGINAL)
type Blockchain struct {
	blocks []*Block
	mu     sync.RWMutex // NUEVO: mutex para acceso concurrente
}

// MiningResult (ORIGINAL)
type MiningResult struct {
	minerName string
	nonce     int
	hash      []byte
}

// MinerStats (ORIGINAL)
type MinerStats struct {
	attempts int
	elapsed  time.Duration
}

// ValidatorInfo contiene información extendida de cada validador
type ValidatorInfo struct {
	Name              string    `json:"name"`
	Stake             int       `json:"stake"`
	BlocksValidated   int       `json:"blocksValidated"`
	AverageResponseMs int       `json:"averageResponseMs"`
	FailureCount      int       `json:"failureCount"`
	MLScore           float64   `json:"mlScore"`
	LastValidation    time.Time `json:"lastValidation"`
}

// MLPredictionRequest es lo que enviamos al servicio de ML
type MLPredictionRequest struct {
	Validators []MLValidatorFeatures `json:"validators"`
}

// MLValidatorFeatures son las características que usa el ML
type MLValidatorFeatures struct {
	Name              string  `json:"name"`
	Stake             float64 `json:"stake"`
	BlocksValidated   float64 `json:"blocks_validated"`
	AverageResponseMs float64 `json:"avg_response_ms"`
	FailureCount      float64 `json:"failure_count"`
}

// MLPredictionResponse es lo que recibimos del servicio de ML
type MLPredictionResponse struct {
	Predictions []MLPrediction `json:"predictions"`
}

// MLPrediction contiene el score de cada validador
type MLPrediction struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

// ================================================================
// ESTADO GLOBAL
// ================================================================

var (
	blockchain *Blockchain

	// stakeTable original (MANTENER)
	stakeTable = map[string]int{
		"Alice":   35,
		"Bob":     32,
		"Charlie": 28,
		"David":   25,
	}

	// Información extendida de validadores
	validatorInfo = map[string]*ValidatorInfo{
		"Alice":   {Name: "Alice", Stake: 35, AverageResponseMs: 80, FailureCount: 2},
		"Bob":     {Name: "Bob", Stake: 32, AverageResponseMs: 78, FailureCount: 1},
		"Charlie": {Name: "Charlie", Stake: 28, AverageResponseMs: 75, FailureCount: 2},
		"David":   {Name: "David", Stake: 25, AverageResponseMs: 82, FailureCount: 1},
	}

	minerList = []string{"Miner1", "Miner2", "Miner3", "Miner4"}

	// URL del servicio de ML
	mlServiceURL = "http://ml-service:5000"
)

// ================================================================
// FUNCIONES DE HASH (ORIGINALES)
// ================================================================

func calculateHash(data []byte, prevHash []byte, nonce int, validator string) []byte {
	allFields := bytes.Join(
		[][]byte{
			data,
			prevHash,
			[]byte(strconv.Itoa(nonce)),
			[]byte(validator),
		},
		[]byte{},
	)

	result := sha256.Sum256(allFields)
	return result[:]
}

func (b *Block) DeriveHash() {
	b.Hash = calculateHash(b.Data, b.PrevHash, b.Nonce, b.Validator)
}

// ================================================================
// PROOF OF STAKE (ORIGINAL)
// ================================================================

func totalStake() int {
	total := 0
	for _, stake := range stakeTable {
		total += stake
	}
	return total
}

func selectValidatorByStake() string {
	total := totalStake()
	randomNumber := rand.Intn(total)

	accumulated := 0
	for name, stake := range stakeTable {
		accumulated += stake
		if randomNumber < accumulated {
			return name
		}
	}

	return ""
}

// ================================================================
// MACHINE LEARNING INTEGRATION (NUEVO)
// ================================================================

// getMLScores llama al servicio de ML y obtiene scores para todos los validadores
func getMLScores() (map[string]float64, error) {
	// Preparar features
	var features []MLValidatorFeatures
	for name, info := range validatorInfo {
		features = append(features, MLValidatorFeatures{
			Name:              name,
			Stake:             float64(info.Stake),
			BlocksValidated:   float64(info.BlocksValidated),
			AverageResponseMs: float64(info.AverageResponseMs),
			FailureCount:      float64(info.FailureCount),
		})
	}

	requestBody := MLPredictionRequest{Validators: features}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// Llamar al servicio de ML
	resp, err := http.Post(mlServiceURL+"/predict", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("⚠️  ML service not available, falling back to PoS: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mlResponse MLPredictionResponse
	if err := json.Unmarshal(body, &mlResponse); err != nil {
		return nil, err
	}

	// Convertir a mapa
	scores := make(map[string]float64)
	for _, pred := range mlResponse.Predictions {
		scores[pred.Name] = pred.Score
	}

	return scores, nil
}

// selectValidatorByML elige el validador con mejor score de ML
func selectValidatorByML() (string, float64, error) {
	scores, err := getMLScores()
	if err != nil {
		return "", 0, err
	}

	// Sumar todos los scores
	total := 0.0
	for _, score := range scores {
		total += score
	}

	// Número aleatorio
	r := rand.Float64() * total

	// Selección tipo ruleta
	accumulated := 0.0
	for name, score := range scores {
		accumulated += score
		if r <= accumulated {
			return name, score, nil
		}
	}

	return "", 0, fmt.Errorf("no validator selected")
}

// ================================================================
// PROOF OF WORK (ORIGINAL - SIMPLIFICADO PARA API)
// ================================================================

func mineBlock(data []byte, prevHash []byte, difficulty int) MiningResult {
	target := bytes.Repeat([]byte{0}, difficulty)
	resultChan := make(chan MiningResult, 1)
	stopChan := make(chan struct{})

	var mu sync.Mutex
	statsMap := make(map[string]*MinerStats)
	for _, name := range minerList {
		statsMap[name] = &MinerStats{}
	}

	for _, name := range minerList {
		go func(minerName string) {
			startTime := time.Now()
			attempts := 0

			for {
				select {
				case <-stopChan:
					mu.Lock()
					statsMap[minerName].attempts = attempts
					statsMap[minerName].elapsed = time.Since(startTime)
					mu.Unlock()
					return
				default:
				}

				attempts++
				nonce := rand.Intn(100_000_000)
				hash := calculateHash(data, prevHash, nonce, "")

				if bytes.HasPrefix(hash, target) {
					mu.Lock()
					statsMap[minerName].attempts = attempts
					statsMap[minerName].elapsed = time.Since(startTime)
					mu.Unlock()

					resultChan <- MiningResult{
						minerName: minerName,
						nonce:     nonce,
						hash:      hash,
					}
					return
				}
			}
		}(name)
	}

	winner := <-resultChan
	close(stopChan)
	time.Sleep(15 * time.Millisecond)

	log.Printf("⛏️  Block mined by %s after %d attempts", winner.minerName, statsMap[winner.minerName].attempts)

	return winner
}

// ================================================================
// CREAR BLOQUE (EXTENDIDO)
// ================================================================

func createBlock(record *AuditRecord, prevHash []byte, consensus string) *Block {
	// Serializar el registro a JSON para almacenarlo en Data
	dataBytes, _ := json.Marshal(record)

	block := &Block{
		Data:      dataBytes,
		PrevHash:  prevHash,
		Timestamp: time.Now(),
		Record:    record,
	}

	switch consensus {
	case "pow":
		result := mineBlock(block.Data, block.PrevHash, 2)
		block.Validator = result.minerName
		block.Nonce = result.nonce
		block.Hash = result.hash
		block.MLScore = 0 // PoW no usa ML

	case "pos":
		chosen := selectValidatorByStake()
		block.Validator = chosen
		block.DeriveHash()
		block.MLScore = 0 // PoS tradicional no usa ML

	case "ml":
		// NUEVO: Selección basada en Machine Learning
		chosen, score, err := selectValidatorByML()
		if err != nil {
			// Fallback a PoS si ML falla
			log.Printf("⚠️  ML selection failed, using PoS: %v", err)
			chosen = selectValidatorByStake()
			score = 0
		}
		block.Validator = chosen
		block.MLScore = score
		block.DeriveHash()

		// Actualizar estadísticas del validador
		if info, ok := validatorInfo[chosen]; ok {
			info.BlocksValidated++
			info.LastValidation = time.Now()
			info.MLScore = score
		}
	}

	return block
}

// ================================================================
// BLOCKCHAIN (EXTENDIDO)
// ================================================================

func (chain *Blockchain) addBlock(record *AuditRecord, consensus string) {
	chain.mu.Lock()
	defer chain.mu.Unlock()

	lastBlock := chain.blocks[len(chain.blocks)-1]
	newBlock := createBlock(record, lastBlock.Hash, consensus)
	chain.blocks = append(chain.blocks, newBlock)
}

func initBlockchain(consensus string) *Blockchain {
	genesisRecord := &AuditRecord{
		Type:        "system",
		UserID:      "system",
		Action:      "genesis",
		Description: "Genesis block - blockchain initialized",
		Timestamp:   time.Now(),
		Metadata:    map[string]interface{}{"version": "1.0"},
	}

	genesis := createBlock(genesisRecord, []byte{}, consensus)
	return &Blockchain{blocks: []*Block{genesis}}
}

func (chain *Blockchain) getBlocks() []*Block {
	chain.mu.RLock()
	defer chain.mu.RUnlock()

	// Crear copia para evitar race conditions
	blocks := make([]*Block, len(chain.blocks))
	copy(blocks, chain.blocks)
	return blocks
}

// ================================================================
// API REST HANDLERS
// ================================================================

// POST /record - Agregar nuevo registro a la blockchain
func handleAddRecord(w http.ResponseWriter, r *http.Request) {
	var record AuditRecord
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	record.Timestamp = time.Now()

	// Obtener consenso de query params (default: ml)
	consensus := r.URL.Query().Get("consensus")
	if consensus == "" {
		consensus = "ml"
	}

	blockchain.addBlock(&record, consensus)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Record added to blockchain",
		"record":  record,
	})

	log.Printf("✅ New record added: %s - %s", record.Type, record.Action)
}

// GET /blockchain - Obtener todos los bloques
func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	blocks := blockchain.getBlocks()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"count":  len(blocks),
		"blocks": blocks,
	})
}

// GET /validators - Obtener información de validadores
func handleGetValidators(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"validators": validatorInfo,
	})
}

// GET /validator-scores - Obtener scores de ML
func handleGetValidatorScores(w http.ResponseWriter, r *http.Request) {
	scores, err := getMLScores()
	if err != nil {
		http.Error(w, "ML service unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"scores": scores,
	})
}

// POST /simulate - Generar datos de prueba
func handleSimulate(w http.ResponseWriter, r *http.Request) {
	types := []string{"log", "audit", "event"}
	actions := []string{"login", "logout", "update", "delete", "create", "view"}
	users := []string{"user1", "user2", "user3", "admin"}

	count := 5
	if countParam := r.URL.Query().Get("count"); countParam != "" {
		if c, err := strconv.Atoi(countParam); err == nil {
			count = c
		}
	}

	consensus := r.URL.Query().Get("consensus")
	if consensus == "" {
		consensus = "ml"
	}

	for i := 0; i < count; i++ {
		record := &AuditRecord{
			Type:        types[rand.Intn(len(types))],
			UserID:      users[rand.Intn(len(users))],
			Action:      actions[rand.Intn(len(actions))],
			Description: fmt.Sprintf("Simulated action %d", i+1),
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"ip":        fmt.Sprintf("192.168.1.%d", rand.Intn(255)),
				"sessionId": fmt.Sprintf("sess_%d", rand.Intn(1000)),
			},
		}

		blockchain.addBlock(record, consensus)
		time.Sleep(100 * time.Millisecond) // Pequeña pausa entre bloques
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("%d records simulated", count),
	})

	log.Printf("🎲 Generated %d simulated records", count)
}

// GET /stats - Estadísticas generales
func handleGetStats(w http.ResponseWriter, r *http.Request) {
	blocks := blockchain.getBlocks()

	// Contar por tipo
	typeCounts := make(map[string]int)
	validatorCounts := make(map[string]int)

	for _, block := range blocks {
		if block.Record != nil {
			typeCounts[block.Record.Type]++
		}
		validatorCounts[block.Validator]++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "success",
		"totalBlocks":     len(blocks),
		"typeDistribution": typeCounts,
		"validatorDistribution": validatorCounts,
	})
}

// ================================================================
// CORS MIDDLEWARE
// ================================================================

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ================================================================
// MAIN
// ================================================================

func main() {
	rand.Seed(time.Now().UnixNano())

	// Inicializar blockchain con consenso ML por defecto
	blockchain = initBlockchain("ml")

	// Configurar router
	router := mux.NewRouter()

	// Endpoints
	router.HandleFunc("/record", handleAddRecord).Methods("POST", "OPTIONS")
	router.HandleFunc("/blockchain", handleGetBlockchain).Methods("GET", "OPTIONS")
	router.HandleFunc("/validators", handleGetValidators).Methods("GET", "OPTIONS")
	router.HandleFunc("/validator-scores", handleGetValidatorScores).Methods("GET", "OPTIONS")
	router.HandleFunc("/simulate", handleSimulate).Methods("POST", "OPTIONS")
	router.HandleFunc("/stats", handleGetStats).Methods("GET", "OPTIONS")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}).Methods("GET")

	// Aplicar CORS
	handler := enableCORS(router)

	// Servidor
	port := ":8080"
	log.Printf("🚀 Blockchain API Server started on http://localhost%s", port)
	log.Printf("📊 Endpoints:")
	log.Printf("   POST   /record              - Add new audit record")
	log.Printf("   GET    /blockchain          - Get all blocks")
	log.Printf("   GET    /validators          - Get validator info")
	log.Printf("   GET    /validator-scores    - Get ML scores")
	log.Printf("   POST   /simulate            - Generate test data")
	log.Printf("   GET    /stats               - Get statistics")
	log.Printf("   GET    /health              - Health check")

	log.Fatal(http.ListenAndServe(port, handler))
}