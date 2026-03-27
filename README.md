# 🔐 Plataforma de Auditoría y Validación con Blockchain + Machine Learning

Sistema completo de auditoría inmutable que combina **Blockchain**, **Machine Learning** y una arquitectura **SaaS** moderna.

## 📋 Descripción

Este proyecto extiende una implementación base de blockchain en Go para crear una plataforma empresarial completa que:

- ✅ Registra eventos de auditoría en una blockchain inmutable
- 🤖 Usa Machine Learning para seleccionar validadores óptimos
- 📊 Proporciona un dashboard web interactivo
- 🚀 Despliega fácilmente con Docker Compose

## 🏗️ Arquitectura

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (Web)                        │
│                    Nginx + HTML/CSS/JS                       │
│                      Puerto: 3000                            │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ HTTP REST API
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     Backend (Go)                             │
│              API REST + Blockchain Logic                     │
│                      Puerto: 8080                            │
│                                                              │
│  • POST /record          - Agregar registro                 │
│  • GET  /blockchain      - Ver blockchain                   │
│  • GET  /validators      - Info validadores                 │
│  • GET  /validator-scores - Scores ML                       │
│  • POST /simulate        - Generar datos                    │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ HTTP Request
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  ML Service (Python)                         │
│           Flask + Scikit-learn (Random Forest)              │
│                      Puerto: 5000                            │
│                                                              │
│  • POST /predict   - Predecir scores validadores            │
│  • POST /retrain   - Re-entrenar modelo                     │
│  • GET  /info      - Información del modelo                 │
└─────────────────────────────────────────────────────────────┘
```

## 🎯 Características Principales

### 1. Blockchain Inmutable
- Implementación completa en Go basada en tu código original
- Soporte para múltiples mecanismos de consenso:
  - **PoW** (Proof of Work) - Minería competitiva
  - **PoS** (Proof of Stake) - Selección por stake
  - **ML** (Machine Learning) - Selección inteligente ⭐

### 2. Machine Learning para Validadores
El modelo evalúa validadores considerando:
- 💰 **Stake**: Cantidad apostada
- 📊 **Historial**: Bloques validados previamente
- ⚡ **Rendimiento**: Tiempo de respuesta promedio
- ❌ **Fiabilidad**: Tasa de fallos

### 3. Tipos de Registros
- **Logs**: Eventos del sistema
- **Auditorías**: Acciones críticas
- **Eventos**: Notificaciones y alertas

### 4. API REST Completa
Todos los endpoints documentados y listos para integración.

## 🚀 Instalación y Ejecución

### Prerrequisitos
- Docker
- Docker Compose

### Paso 1: Clonar el proyecto
```bash
# Ya tienes los archivos en blockchain-ml-saas/
cd blockchain-ml-saas
```

### Paso 2: Iniciar todos los servicios
```bash
docker-compose up --build
```

Esto levantará:
- ✅ **ML Service** en http://localhost:5000
- ✅ **Backend API** en http://localhost:8080
- ✅ **Frontend** en http://localhost:3000

### Paso 3: Abrir el Dashboard
Navega a: **http://localhost:3000**

## 📖 Uso del Sistema

### Desde el Dashboard Web

1. **Agregar Registro Manual**
   - Selecciona tipo (log/audit/event)
   - Ingresa usuario y acción
   - Elige mecanismo de consenso
   - Click en "Agregar Registro"

2. **Simular Datos**
   - Click en "Simular Datos" para generar 5 registros automáticos

3. **Visualizar Blockchain**
   - Los últimos 10 bloques se muestran automáticamente
   - Se actualiza cada 5 segundos

4. **Ver Validadores**
   - Muestra scores de ML en tiempo real
   - Compara rendimiento entre validadores

### Desde la API (cURL)

```bash
# Agregar un registro
curl -X POST http://localhost:8080/record?consensus=ml \
  -H "Content-Type: application/json" \
  -d '{
    "type": "audit",
    "userId": "admin",
    "action": "delete_user",
    "description": "Usuario eliminado por política de seguridad"
  }'

# Ver blockchain completa
curl http://localhost:8080/blockchain | jq

# Ver validadores
curl http://localhost:8080/validators | jq

# Ver scores de ML
curl http://localhost:8080/validator-scores | jq

# Simular 10 registros
curl -X POST "http://localhost:8080/simulate?count=10&consensus=ml"

# Ver estadísticas
curl http://localhost:8080/stats | jq
```

## 🤖 Cómo Funciona el Machine Learning

### Entrenamiento
El modelo se entrena con datos sintéticos que simulan tres tipos de validadores:

1. **Validadores Buenos** (Score alto)
   - Alto stake (≥40 tokens)
   - Mucha experiencia (≥90 validaciones)
   - Rápidos (≤90ms)
   - Sin fallos (≤1)

2. **Validadores Promedio** (Score medio)
   - Stake moderado (~30 tokens)
   - Experiencia media (~50 validaciones)
   - Velocidad normal (~120ms)
   - Pocos fallos (~2-3)

3. **Validadores Débiles** (Score bajo)
   - Bajo stake (≤15 tokens)
   - Poca experiencia (≤20 validaciones)
   - Lentos (≥180ms)
   - Muchos fallos (≥7)

### Predicción
```python
# Características de entrada
features = [stake, blocks_validated, avg_response_ms, failure_count]

# El modelo devuelve un score entre 0 y 1
score = model.predict(features)

# El validador con mayor score es elegido
```

### Ventajas sobre PoS Tradicional

| Aspecto | PoS Tradicional | ML-Based |
|---------|----------------|----------|
| Criterio | Solo stake | Múltiples factores |
| Rendimiento | No considerado | ✅ Evaluado |
| Historial | No considerado | ✅ Evaluado |
| Confiabilidad | No considerado | ✅ Evaluado |
| Equidad | Favorece ricos | ✅ Más justo |

## 📊 Estructura del Código

### Backend (Go)
```
backend/
├── main.go              # Código principal extendido
├── go.mod               # Dependencias
└── Dockerfile           # Imagen Docker
```

**Extensiones sobre tu código original:**
- ✅ Tipos de datos estructurados (`AuditRecord`)
- ✅ API REST con Gorilla Mux
- ✅ Integración con servicio de ML
- ✅ Nuevo consenso "ml"
- ✅ Estadísticas de validadores
- ✅ Thread-safety con mutex

### ML Service (Python)
```
ml-service/
├── app.py               # Flask API + modelo ML
├── requirements.txt     # Dependencias Python
└── Dockerfile           # Imagen Docker
```

**Componentes:**
- `ValidatorScorer`: Clase del modelo
- Random Forest Classifier (100 árboles)
- Endpoints para predicción y re-entrenamiento

### Frontend (Web)
```
frontend/
├── index.html           # Dashboard completo
└── Dockerfile           # Nginx para servir HTML
```

**Características:**
- Diseño responsive
- Auto-refresh cada 5 segundos
- Visualización de blockchain
- Comparación ML vs PoS

## 🔧 Configuración Avanzada

### Cambiar URL del ML Service

En `backend/main.go`:
```go
mlServiceURL = "http://ml-service:5000"  // Cambiar aquí
```

### Re-entrenar el Modelo

```bash
curl -X POST http://localhost:5000/retrain \
  -H "Content-Type: application/json" \
  -d '{
    "training_data": [
      {"features": [50, 100, 80, 0], "quality": 2},
      {"features": [30, 50, 120, 2], "quality": 1}
    ]
  }'
```

### Agregar Nuevos Validadores

En `backend/main.go`:
```go
validatorInfo = map[string]*ValidatorInfo{
    "Alice":   {Name: "Alice", Stake: 50, ...},
    // Agregar nuevos aquí
    "Eve":     {Name: "Eve", Stake: 40, ...},
}

stakeTable = map[string]int{
    "Alice":   50,
    // Agregar nuevos aquí
    "Eve":     40,
}
```

## 📈 Casos de Uso

### 1. Auditoría Empresarial
- Registrar accesos a sistemas críticos
- Trazabilidad de cambios en bases de datos
- Compliance y regulaciones (GDPR, SOX)

### 2. Logs de Aplicación
- Eventos de seguridad
- Errores y excepciones
- Métricas de rendimiento

### 3. Cadena de Custodia
- Documentos legales
- Evidencia digital
- Certificaciones

## 🧪 Testing

### Verificar Servicios
```bash
# ML Service
curl http://localhost:5000/health

# Backend
curl http://localhost:8080/health

# Frontend
curl http://localhost:3000
```

### Probar Consensos
```bash
# Proof of Work
curl -X POST "http://localhost:8080/simulate?consensus=pow&count=1"

# Proof of Stake
curl -X POST "http://localhost:8080/simulate?consensus=pos&count=1"

# Machine Learning
curl -X POST "http://localhost:8080/simulate?consensus=ml&count=1"
```

## 🎓 Explicación Técnica Detallada

### ¿Cómo se Mantiene la Inmutabilidad?

Cada bloque contiene:
1. `Hash`: SHA-256 del bloque actual
2. `PrevHash`: Hash del bloque anterior

Si alguien modifica un bloque:
- Su hash cambia
- El siguiente bloque apunta a un hash incorrecto
- La cadena se rompe → detección inmediata

### ¿Cómo Funciona la Minería (PoW)?

```go
// Los mineros buscan un nonce que produzca un hash válido
for {
    nonce = random()
    hash = SHA256(data + prevHash + nonce)
    
    if hash.startsWith("0000") {  // Dificultad = 2 bytes
        return nonce  // ¡Encontrado!
    }
}
```

### ¿Cómo Selecciona el ML?

```python
# 1. Extraer features de validadores
features = [[stake, validations, response_time, failures], ...]

# 2. Normalizar (StandardScaler)
features_scaled = scaler.transform(features)

# 3. Predecir con Random Forest
scores = model.predict_proba(features_scaled)

# 4. Elegir el de mayor score
best_validator = validators[argmax(scores)]
```

## 🐛 Troubleshooting

### El frontend no carga
- Verifica que los 3 servicios estén corriendo: `docker-compose ps`
- Revisa logs: `docker-compose logs frontend`

### ML Service no responde
- El backend cae automáticamente a PoS si ML falla
- Revisa: `docker-compose logs ml-service`

### Error de CORS
- Asegúrate de acceder desde `localhost`, no desde IP
- El backend ya tiene CORS habilitado

## 🚀 Próximos Pasos (Mejoras Posibles)

1. **Persistencia**: Agregar base de datos (PostgreSQL)
2. **Autenticación**: JWT para usuarios
3. **Websockets**: Updates en tiempo real
4. **Métricas**: Prometheus + Grafana
5. **Tests**: Unit tests y integration tests
6. **Modelo avanzado**: Deep Learning con TensorFlow
7. **Sharding**: Escalar horizontalmente

## 📝 Licencia

Este proyecto es de código abierto. Siéntete libre de usarlo y modificarlo.

## 👨‍💻 Autor

Construido sobre tu implementación base de blockchain en Go.
Extendido con arquitectura SaaS completa.

---

## 🎉 ¡Disfruta tu Blockchain + ML Platform!

Para cualquier duda:
1. Revisa los logs: `docker-compose logs -f`
2. Verifica el dashboard: http://localhost:3000
3. Explora la API: http://localhost:8080

**Happy Hacking! 🚀**