import yaml
import time
import logging
from pathlib import Path
from typing import List, Dict, Any
from datetime import datetime

import numpy as np
from sentence_transformers import SentenceTransformer
import chromadb
from sklearn.cluster import DBSCAN
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler
from rich.console import Console
from rich.table import Table
from rich import print as rprint

class LogProcessor:
    def __init__(self, config_path: str = "config.yaml"):
        with open(config_path, 'r') as f:
            self.config = yaml.safe_load(f)
        
        # Initialize components
        self.console = Console()
        self.model = SentenceTransformer(
            self.config['model']['name'],
            device=self.config['model']['device']
        )
        
        # Initialize ChromaDB (persistent client)
        self.chroma_client = chromadb.PersistentClient(path="./chroma_db")
        self.collection = self.chroma_client.get_or_create_collection(
            name="log_embeddings",
            metadata={"description": "Log message embeddings for anomaly detection"}
        )
        
        self.processed_count = 0
        self.anomaly_count = 0
        
        self.console.print("✅ [green]Log Processor initialized[/green]")
        self.console.print(f"🤖 Using model: {self.config['model']['name']}")
    
    def embed_log_message(self, log_message: str) -> List[float]:
        """Convert log message to vector embedding"""
        return self.model.encode([log_message])[0].tolist()
    
    def process_log_line(self, log_line: str):
        """Process a single log line"""
        if not log_line.strip():
            return
        
        try:
            # Create embedding
            embedding = self.embed_log_message(log_line)
            
            # Store in ChromaDB
            log_id = f"log_{int(time.time() * 1000)}_{self.processed_count}"
            self.collection.add(
                embeddings=[embedding],
                documents=[log_line],
                metadatas=[{
                    "timestamp": datetime.now().isoformat(),
                    "processed_at": time.time()
                }],
                ids=[log_id]
            )
            
            self.processed_count += 1
            
            # Run analysis periodically
            if self.processed_count % self.config['processing']['batch_size'] == 0:
                self.analyze_anomalies()
                
        except Exception as e:
            self.console.print(f"❌ [red]Error processing log: {e}[/red]")
    
    def analyze_anomalies(self):
        """Analyze recent logs for anomalies using clustering"""
        try:
            # Get recent embeddings from ChromaDB
            results = self.collection.get(
                limit=self.config['processing']['max_logs_in_memory'],
                include=['embeddings', 'documents', 'metadatas']
            )
            
            if len(results['embeddings']) < self.config['clustering']['min_samples']:
                return
            
            embeddings = np.array(results['embeddings'])
            documents = results['documents']
            
            # Perform DBSCAN clustering
            clustering = DBSCAN(
                eps=self.config['clustering']['eps'],
                min_samples=self.config['clustering']['min_samples']
            ).fit(embeddings)
            
            # Identify anomalies (points labeled as -1 are outliers)
            anomaly_indices = np.where(clustering.labels_ == -1)[0]
            
            if len(anomaly_indices) > 0:
                self.display_anomalies(anomaly_indices, documents, clustering.labels_)
                
        except Exception as e:
            self.console.print(f"❌ [red]Error in anomaly analysis: {e}[/red]")
    
    def display_anomalies(self, anomaly_indices, documents, labels):
        """Display detected anomalies in a nice format"""
        self.anomaly_count += len(anomaly_indices)
        
        table = Table(
            title=f"🚨 Semantic Log Anomalies Detected ({len(anomaly_indices)} found)",
            show_header=True,
            header_style="bold red"
        )
        table.add_column("Index", style="cyan")
        table.add_column("Log Message", style="white")
        table.add_column("Cluster", style="yellow")
        
        for idx in anomaly_indices:
            table.add_row(
                str(idx),
                documents[idx][:100] + "..." if len(documents[idx]) > 100 else documents[idx],
                str(labels[idx])
            )
        
        self.console.print(table)
        self.console.print(f"📊 [yellow]Total anomalies detected: {self.anomaly_count}[/yellow]")

class LogFileHandler(FileSystemEventHandler):
    def __init__(self, processor: LogProcessor):
        self.processor = processor
    
    def on_modified(self, event):
        if event.is_file and event.src_path.endswith('.log'):
            self.process_new_logs(event.src_path)
    
    def process_new_logs(self, file_path: str):
        """Read new lines from log file"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                # Simple approach: process all lines
                # In production, you'd track file position
                for line in f:
                    if line.strip():
                        self.processor.process_log_line(line.strip())
        except Exception as e:
            print(f"Error reading log file: {e}")

def main():
    """Main execution function"""
    # Create sample logs directory
    Path("./sample_logs").mkdir(exist_ok=True)
    
    # Initialize processor
    processor = LogProcessor()
    
    # Set up file watcher
    event_handler = LogFileHandler(processor)
    observer = Observer()
    observer.schedule(
        event_handler,
        processor.config['files']['watch_directory'],
        recursive=False
    )
    observer.start()
    
    processor.console.print("👀 [blue]Starting log file monitoring...[/blue]")
    processor.console.print(f"📁 Watching: {processor.config['files']['watch_directory']}")
    
    try:
        # Process existing logs first
        log_file = processor.config['files']['log_path']
        if Path(log_file).exists():
            processor.console.print("📄 Processing existing log file...")
            with open(log_file, 'r') as f:
                for line in f:
                    processor.process_log_line(line.strip())
        
        # Keep running
        while True:
            time.sleep(processor.config['processing']['analysis_interval_seconds'])
            processor.analyze_anomalies()
            
    except KeyboardInterrupt:
        processor.console.print("\n🛑 [yellow]Shutting down log analyzer...[/yellow]")
        observer.stop()
    
    observer.join()

if __name__ == "__main__":
    main()
