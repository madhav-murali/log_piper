import random
import time
from datetime import datetime
from pathlib import Path

# Normal log patterns
NORMAL_LOGS = [
    "GET /api/users 200 45ms",
    "POST /api/login 200 120ms", 
    "GET /api/products 200 23ms",
    "PUT /api/users/123 200 67ms",
    "DELETE /api/sessions 200 89ms",
    "GET /api/health 200 12ms",
    "POST /api/orders 201 156ms"
]

# Anomalous log patterns (these should be detected as outliers)
ANOMALOUS_LOGS = [
    "CRITICAL: Database connection pool exhausted",
    "SECURITY ALERT: Multiple failed login attempts from IP 192.168.1.100",
    "Memory leak detected: Heap usage at 95%",
    "Unusual query pattern: SELECT * FROM users WHERE 1=1",
    "ERROR: Payment gateway timeout after 30s",
    "FATAL: Kernel panic - not syncing",
    "ALERT: Cross-site scripting attempt detected"
]

def generate_logs():
    """Generate sample log data with occasional anomalies"""
    log_file = Path("./sample_logs/app.log")
    log_file.parent.mkdir(exist_ok=True)
    
    count = 0
    print("Generating sample logs... Press Ctrl+C to stop.")
    
    try:
        with open(log_file, 'w') as f:
            while True:
                # 95% normal logs, 5% anomalies
                if random.random() < 0.95:
                    log = random.choice(NORMAL_LOGS)
                else:
                    log = random.choice(ANOMALOUS_LOGS)
                
                timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
                log_line = f"{timestamp} - {log}\n"
                
                f.write(log_line)
                f.flush()  # Ensure it's written immediately
                
                count += 1
                print(f"Generated {count} logs...", end='\r')
                
                time.sleep(0.1)  # Generate ~10 logs per second
                
    except KeyboardInterrupt:
        print(f"\nGenerated {count} total log entries to {log_file}")

if __name__ == "__main__":
    generate_logs()
