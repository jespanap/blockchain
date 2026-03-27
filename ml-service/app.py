"""
Servicio de Machine Learning para selección de validadores
Usa características como stake, historial y rendimiento para predecir
la calidad de cada validador
"""

from flask import Flask, request, jsonify
from flask_cors import CORS
import numpy as np
from sklearn.ensemble import RandomForestClassifier
from sklearn.preprocessing import StandardScaler
import logging

app = Flask(__name__)
CORS(app)

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# ================================================================
# MODELO DE MACHINE LEARNING
# ================================================================

class ValidatorScorer:
    """
    Modelo que puntúa validadores basándose en sus características
    
    Features:
    - stake: Cantidad apostada (mayor = mejor)
    - blocks_validated: Historial de validaciones (mayor = mejor)
    - avg_response_ms: Tiempo de respuesta promedio (menor = mejor)
    - failure_count: Número de fallos (menor = mejor)
    """
    
    def __init__(self):
        self.scaler = StandardScaler()
        self.model = None
        self.is_trained = False
        
        # Entrenar con datos sintéticos al inicializar
        self._train_with_synthetic_data()
    
    def _train_with_synthetic_data(self):
        """
        Entrena el modelo con datos simulados
        Simula diferentes perfiles de validadores
        """
        logger.info("🎓 Training ML model with synthetic data...")
        
        # Generar datos de entrenamiento
        # Formato: [stake, blocks_validated, avg_response_ms, failure_count]
        
        # Validadores "buenos" (alta calidad)
        good_validators = np.array([
            [50, 100, 80, 0],   # Muy bueno: alto stake, mucha experiencia, rápido, sin fallos
            [45, 95, 90, 1],
            [40, 90, 85, 0],
            [48, 105, 75, 1],
            [52, 110, 70, 0],
            [47, 98, 88, 1],
        ])
        
        # Validadores "promedio"
        average_validators = np.array([
            [30, 50, 120, 2],   # Promedio: stake medio, experiencia media
            [28, 55, 115, 3],
            [32, 48, 125, 2],
            [29, 52, 118, 3],
            [31, 49, 122, 2],
        ])
        
        # Validadores "débiles" (baja calidad)
        weak_validators = np.array([
            [10, 20, 180, 8],   # Débil: bajo stake, poca experiencia, lento, muchos fallos
            [8, 15, 200, 10],
            [12, 18, 190, 9],
            [9, 22, 185, 7],
            [11, 19, 195, 8],
            [5, 10, 220, 12],
        ])
        
        # Combinar datos
        X_train = np.vstack([good_validators, average_validators, weak_validators])
        
        # Labels: 2 = bueno, 1 = promedio, 0 = débil
        y_train = np.array([2]*len(good_validators) + 
                          [1]*len(average_validators) + 
                          [0]*len(weak_validators))
        
        # Normalizar features
        X_scaled = self.scaler.fit_transform(X_train)
        
        # Entrenar Random Forest
        self.model = RandomForestClassifier(
            n_estimators=100,
            max_depth=10,
            random_state=42
        )
        self.model.fit(X_scaled, y_train)
        
        self.is_trained = True
        
        # Mostrar feature importance
        feature_names = ['stake', 'blocks_validated', 'avg_response_ms', 'failure_count']
        importance = self.model.feature_importances_
        
        logger.info("✅ Model trained successfully!")
        logger.info("📊 Feature Importance:")
        for name, imp in zip(feature_names, importance):
            logger.info(f"   {name}: {imp:.3f}")
    
    def predict_score(self, validators):
        """
        Predice el score de calidad para cada validador
        
        Args:
            validators: Lista de diccionarios con features
            
        Returns:
            Lista de scores normalizados entre 0 y 1
        """
        if not self.is_trained:
            raise ValueError("Model not trained")
        
        # Extraer features
        features = []
        for v in validators:
            features.append([
                v['stake'],
                v['blocks_validated'],
                v['avg_response_ms'],
                v['failure_count']
            ])
        
        X = np.array(features)
        X_scaled = self.scaler.transform(X)
        
        # Obtener probabilidades de cada clase
        probas = self.model.predict_proba(X_scaled)
        
        # Calcular score ponderado: 
        # score = 0.0 * P(clase_0) + 0.5 * P(clase_1) + 1.0 * P(clase_2)
        scores = probas[:, 0] * 0.0 + probas[:, 1] * 0.5 + probas[:, 2] * 1.0
        
        # Añadir pequeño bonus por stake (incentivo económico)
        stake_bonus = np.array([v['stake'] for v in validators]) / 100 * 0.1
        scores = scores + stake_bonus
        
        # Normalizar entre 0 y 1
        scores = np.clip(scores, 0, 1)
        
        return scores.tolist()

# Instancia global del scorer
scorer = ValidatorScorer()

# ================================================================
# API ENDPOINTS
# ================================================================

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({
        'status': 'healthy',
        'model_trained': scorer.is_trained
    })

@app.route('/predict', methods=['POST'])
def predict():
    """
    Predice scores para validadores
    
    Request body:
    {
        "validators": [
            {
                "name": "Alice",
                "stake": 50,
                "blocks_validated": 100,
                "avg_response_ms": 80,
                "failure_count": 0
            },
            ...
        ]
    }
    
    Response:
    {
        "predictions": [
            {"name": "Alice", "score": 0.95},
            ...
        ]
    }
    """
    try:
        data = request.get_json()
        validators = data.get('validators', [])
        
        if not validators:
            return jsonify({'error': 'No validators provided'}), 400
        
        # Predecir scores
        scores = scorer.predict_score(validators)
        
        # Formatear respuesta
        predictions = [
            {
                'name': v['name'],
                'score': round(score, 4)
            }
            for v, score in zip(validators, scores)
        ]
        
        logger.info(f"🤖 Predicted scores for {len(predictions)} validators")
        
        return jsonify({
            'predictions': predictions
        })
        
    except Exception as e:
        logger.error(f"❌ Prediction error: {str(e)}")
        return jsonify({'error': str(e)}), 500

@app.route('/retrain', methods=['POST'])
def retrain():
    """
    Re-entrena el modelo con nuevos datos (opcional)
    
    Request body:
    {
        "training_data": [
            {
                "features": [50, 100, 80, 0],
                "quality": 2  // 0=débil, 1=promedio, 2=bueno
            },
            ...
        ]
    }
    """
    try:
        data = request.get_json()
        training_data = data.get('training_data', [])
        
        if not training_data:
            # Re-entrenar con datos sintéticos
            scorer._train_with_synthetic_data()
            return jsonify({
                'status': 'success',
                'message': 'Model retrained with synthetic data'
            })
        
        # Extraer features y labels
        X = np.array([d['features'] for d in training_data])
        y = np.array([d['quality'] for d in training_data])
        
        # Normalizar y entrenar
        X_scaled = scorer.scaler.fit_transform(X)
        scorer.model.fit(X_scaled, y)
        
        logger.info(f"🎓 Model retrained with {len(training_data)} samples")
        
        return jsonify({
            'status': 'success',
            'message': f'Model retrained with {len(training_data)} samples'
        })
        
    except Exception as e:
        logger.error(f"❌ Retraining error: {str(e)}")
        return jsonify({'error': str(e)}), 500

@app.route('/info', methods=['GET'])
def info():
    """Información sobre el modelo"""
    if not scorer.is_trained:
        return jsonify({'error': 'Model not trained'}), 400
    
    return jsonify({
        'model_type': 'RandomForestClassifier',
        'features': ['stake', 'blocks_validated', 'avg_response_ms', 'failure_count'],
        'feature_importance': {
            'stake': float(scorer.model.feature_importances_[0]),
            'blocks_validated': float(scorer.model.feature_importances_[1]),
            'avg_response_ms': float(scorer.model.feature_importances_[2]),
            'failure_count': float(scorer.model.feature_importances_[3])
        },
        'n_estimators': scorer.model.n_estimators,
        'max_depth': scorer.model.max_depth
    })

# ================================================================
# MAIN
# ================================================================

if __name__ == '__main__':
    logger.info("🚀 Starting ML Service...")
    logger.info("📊 Model will predict validator quality based on:")
    logger.info("   - Stake amount")
    logger.info("   - Validation history")
    logger.info("   - Response time")
    logger.info("   - Failure count")
    
    app.run(host='0.0.0.0', port=5000, debug=False)