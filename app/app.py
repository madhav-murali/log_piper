from fastapi import FastAPI, BackgroundTasks
from pydantic import BaseModel
from typing import List, Optional
import httpx
import os
import json
from datetime import datetime
import logging

# Configuration
HF_API_KEY = os.getenv("HF_API_KEY", "your_hf_key")
HF_MODEL = "sentence-transformers/all-MiniLM-L6-v2"
DATABASE_URL = os.getenv("DATABASE_URL", "sqlite:///./logs.db")

app = FastAPI(title="AIOps Log Analyzer")

class LogEntry(BaseModel):
    message: str
    timestamp: str
    source: str = "local"
    level: str = "INFO"

class AnalysisResult(BaseModel):
    log_id: int
    anomaly_score: float
    is_anomaly: bool
    cluster_label: Optional[int] = None
    processed_at: str

async def get_embedding(text: str) -> List[float]:
    """Get embeddings from Hugging Face Inference API"""
    async with httpx.AsyncClient() as client:
        try:
            response = await client.post(
                f"https://api-inference.huggingface.co/models/{HF_MODEL}",
                headers={"Authorization": f"Bearer {HF_API_KEY}"},
                json={"inputs": text}
            )
            
            response.raise_for_status()  # Check for HTTP errors
            
            result = response.json()
            
            # --- THIS IS THE FIX ---
            # Check if it's a list and not empty
            if isinstance(result, list) and len(result) > 0:
                return result[0]  # Return the first embedding vector
            else:
                logging.error(f"HF API returned unexpected data: {result}")
                return []
        
        except httpx.HTTPStatusError as e:
            logging.error(f"HF API HTTP error: {e.response.status_code} - {e.response.text}")
            return []
        except Exception as e:
            logging.error(f"Error in get_embedding: {e}")
            return []
async def detect_anomaly_simple(embedding: List[float], recent_embeddings: List[List[float]]) -> dict:
    """Simple anomaly detection using distance threshold"""
    if not recent_embeddings:
        return {"is_anomaly": False, "score": 0.0}
    
    # Calculate average distance to recent embeddings
    import numpy as np
    current = np.array(embedding)
    recent = np.array(recent_embeddings)
    
    distances = np.linalg.norm(recent - current, axis=1)
    avg_distance = np.mean(distances)
    
    # Simple threshold-based anomaly detection
    is_anomaly = avg_distance > 0.8  # Adjust based on your data
    return {"is_anomaly": bool(is_anomaly), "score": float(avg_distance)}

@app.post("/analyze")
async def analyze_logs(logs: List[LogEntry], background_tasks: BackgroundTasks):
    """Analyze logs for anomalies using Hugging Face embeddings"""
    results = []
    
    for log in logs:
        # Get embedding from Hugging Face
        embedding = await get_embedding(log.message)
        
        if embedding:
            # Simple anomaly detection (you can enhance this)
            anomaly_result = await detect_anomaly_simple(embedding, [])
            
            result = AnalysisResult(
                log_id=len(results) + 1,
                anomaly_score=anomaly_result["score"],
                is_anomaly=anomaly_result["is_anomaly"],
                processed_at=datetime.utcnow().isoformat()
            )
            results.append(result)
            
            # Log anomalies immediately
            if anomaly_result["is_anomaly"]:
                logging.warning(f"🚨 ANOMALY DETECTED: {log.message}")
    
    return {
        "processed": len(logs),
        "anomalies": len([r for r in results if r.is_anomaly]),
        "results": results
    }

@app.get("/health")
async def health():
    return {"status": "healthy", "model": HF_MODEL}

@app.post("/analyze-single")
async def analyze_single(log: LogEntry):
    """Analyze a single log entry"""
    embedding = await get_embedding(log.message)
    
    if embedding:
        anomaly_result = await detect_anomaly_simple(embedding, [])
        return {
            "message": log.message,
            "is_anomaly": anomaly_result["is_anomaly"],
            "confidence": anomaly_result["score"],
            "embedding_length": len(embedding)
        }
    
    return {"error": "Failed to get embedding"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
