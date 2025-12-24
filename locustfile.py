"""
Locust Load Testing for WhatsApp API
Run: locust -f locustfile.py --host=http://localhost:8080
"""

from locust import HttpUser, task, between, TaskSet
import random
import string

class SessionTasks(TaskSet):
    """Tasks for session management endpoints"""
    
    def on_start(self):
        """Initialize test data"""
        # Generate random agent ID for this user
        self.agent_id = f"agent_test_{self.generate_random_string(8)}"
        self.api_key = "secret"  # Default API key from seed
        
    def generate_random_string(self, length=8):
        """Generate random string for unique IDs"""
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))
    
    @task(3)
    def create_session(self):
        """
        Test session creation.
        This will generate QR code but won't wait for scan.
        Tests QR generation performance and DB write speed.
        """
        payload = {
            "agentId": self.agent_id,
            "agentName": f"Test Agent {self.agent_id}",
            "apiKey": self.api_key,
            "langchainUrl": "https://api.example.com"
        }
        
        with self.client.post(
            "/api/v1/sessions/create",
            json=payload,
            catch_response=True,
            name="/sessions/create"
        ) as response:
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    response.success()
                else:
                    response.failure(f"API returned success=false: {data}")
            elif response.status_code == 409:
                # Session already exists - acceptable in load testing
                response.success()
            else:
                response.failure(f"Unexpected status: {response.status_code}")
    
    @task(5)
    def get_session_status(self):
        """Test getting session status"""
        with self.client.get(
            f"/api/v1/sessions/status?agentId={self.agent_id}",
            catch_response=True,
            name="/sessions/status"
        ) as response:
            if response.status_code == 200:
                response.success()
            elif response.status_code == 404:
                # Session not found - acceptable if create hasn't run yet
                response.success()
            else:
                response.failure(f"Unexpected status: {response.status_code}")
    
    @task(2)
    def get_session_detail(self):
        """Test getting detailed session info"""
        with self.client.get(
            f"/api/v1/sessions/detail?agentId={self.agent_id}",
            catch_response=True,
            name="/sessions/detail"
        ) as response:
            if response.status_code in [200, 404]:
                response.success()
            else:
                response.failure(f"Unexpected status: {response.status_code}")
    
    @task(1)
    def reconnect_session(self):
        """Test session reconnection"""
        payload = {"agentId": self.agent_id}
        
        with self.client.post(
            "/api/v1/sessions/reconnect",
            json=payload,
            catch_response=True,
            name="/sessions/reconnect"
        ) as response:
            if response.status_code in [200, 404, 500]:
                # Accept these statuses in load testing
                response.success()
            else:
                response.failure(f"Unexpected status: {response.status_code}")
    
    @task(1)
    def delete_session(self):
        """Test session deletion"""
        payload = {"agentId": self.agent_id}
        
        with self.client.delete(
            "/api/v1/sessions/delete",
            json=payload,
            catch_response=True,
            name="/sessions/delete"
        ) as response:
            if response.status_code in [200, 404]:
                response.success()
            else:
                response.failure(f"Unexpected status: {response.status_code}")


class LangchainTasks(TaskSet):
    """Tasks for Langchain execution endpoint"""
    
    def on_start(self):
        """Initialize test data"""
        self.agent_id = f"agent_test_{self.generate_random_string(8)}"
        self.messages = [
            "Hello, how are you?",
            "What's the weather like?",
            "Tell me a joke",
            "Help me with my task",
            "Thank you!"
        ]
    
    def generate_random_string(self, length=8):
        """Generate random string for unique IDs"""
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))
    
    @task
    def execute_langchain(self):
        """
        Test Langchain execution.
        Note: This will likely fail if session is not connected,
        but tests the endpoint performance.
        """
        payload = {
            "agentId": self.agent_id,
            "message": random.choice(self.messages),
            "sender": f"user_{self.generate_random_string(6)}",
            "params": {
                "max_steps": 5
            }
        }
        
        with self.client.post(
            "/api/v1/langchain/execute",
            json=payload,
            catch_response=True,
            name="/langchain/execute"
        ) as response:
            if response.status_code == 200:
                response.success()
            elif response.status_code in [404, 500]:
                # Expected if session doesn't exist or isn't connected
                response.success()
            else:
                response.failure(f"Unexpected status: {response.status_code}")


class HealthCheckTasks(TaskSet):
    """Tasks for health check endpoint"""
    
    @task
    def health_check(self):
        """Test health endpoint"""
        self.client.get("/health", name="/health")


class WhatsAppAPIUser(HttpUser):
    """
    Simulated user for WhatsApp API.
    Each user will randomly execute tasks from different task sets.
    """
    
    # Wait between 1-3 seconds between tasks
    wait_time = between(1, 3)
    
    # Weight distribution for different task sets
    tasks = {
        SessionTasks: 5,      # 50% session operations
        LangchainTasks: 3,    # 30% langchain executions
        HealthCheckTasks: 2   # 20% health checks
    }


class FocusedSessionUser(HttpUser):
    """
    User that only tests session endpoints.
    Use this for focused session load testing.
    Run with: locust -f locustfile.py --host=http://localhost:8080 FocusedSessionUser
    """
    wait_time = between(0.5, 2)
    tasks = [SessionTasks]


class FocusedLangchainUser(HttpUser):
    """
    User that only tests Langchain endpoint.
    Use this for focused AI integration testing.
    Run with: locust -f locustfile.py --host=http://localhost:8080 FocusedLangchainUser
    """
    wait_time = between(1, 4)
    tasks = [LangchainTasks]
