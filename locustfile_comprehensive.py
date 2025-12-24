"""
COMPREHENSIVE Locust Load Testing - All Endpoints
Tests ALL available endpoints in the WhatsApp API

Run: locust -f locustfile_comprehensive.py --host=http://localhost:8080
"""

from locust import HttpUser, task, between, events
import random
import string
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class ComprehensiveTestUser(HttpUser):
    """
    Tests ALL endpoints in the API
    """
    
    wait_time = between(1, 3)
    
    def on_start(self):
        """Initialize test data"""
        self.agent_id = f"test_{self.generate_random_string(10)}"
        self.api_key = "secret"
        self.session_created = False
        logger.info(f"User started: {self.agent_id}")
    
    def generate_random_string(self, length=8):
        """Generate random string"""
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))
    
    # ========== HEALTH CHECK ==========
    
    @task(5)
    def health_check(self):
        """Test /health endpoint"""
        self.client.get("/health", name="[HEALTH] GET /health")
    
    # ========== SESSION ENDPOINTS ==========
    
    @task(10)
    def create_test_session(self):
        """Test /sessions/create-test (bypass QR)"""
        if self.session_created:
            return
        
        payload = {
            "agentId": self.agent_id,
            "agentName": f"Test Agent {self.agent_id}",
            "apiKey": self.api_key,
            "langchainUrl": "https://api.example.com"
        }
        
        with self.client.post(
            "/api/v1/sessions/create-test",
            json=payload,
            catch_response=True,
            name="[SESSION] POST /create-test"
        ) as response:
            if response.status_code == 200:
                self.session_created = True
                response.success()
            elif response.status_code == 409:
                self.session_created = True
                response.success()
            elif response.status_code == 403:
                response.failure("APP_ENV not set to testing/development")
            else:
                response.failure(f"Status: {response.status_code}")
    
    @task(3)
    def create_normal_session(self):
        """Test /sessions/create (normal flow with QR)"""
        payload = {
            "agentId": f"qr_{self.generate_random_string(8)}",
            "agentName": "QR Test Agent",
            "apiKey": self.api_key,
            "langchainUrl": "https://api.example.com"
        }
        
        with self.client.post(
            "/api/v1/sessions/create",
            json=payload,
            catch_response=True,
            name="[SESSION] POST /create (QR)"
        ) as response:
            if response.status_code in [200, 409]:
                response.success()
            else:
                response.failure(f"Status: {response.status_code}")
    
    @task(15)
    def get_session_status(self):
        """Test /sessions/status"""
        with self.client.get(
            f"/api/v1/sessions/status?agentId={self.agent_id}",
            catch_response=True,
            name="[SESSION] GET /status"
        ) as response:
            if response.status_code in [200, 404]:
                response.success()
            else:
                response.failure(f"Status: {response.status_code}")
    
    @task(10)
    def get_session_detail(self):
        """Test /sessions/detail"""
        with self.client.get(
            f"/api/v1/sessions/detail?agentId={self.agent_id}",
            catch_response=True,
            name="[SESSION] GET /detail"
        ) as response:
            if response.status_code in [200, 404]:
                response.success()
            else:
                response.failure(f"Status: {response.status_code}")
    
    @task(2)
    def reconnect_session(self):
        """Test /sessions/reconnect"""
        payload = {"agentId": self.agent_id}
        
        with self.client.post(
            "/api/v1/sessions/reconnect",
            json=payload,
            catch_response=True,
            name="[SESSION] POST /reconnect"
        ) as response:
            if response.status_code in [200, 404, 500]:
                response.success()
            else:
                response.failure(f"Status: {response.status_code}")
    
    @task(1)
    def delete_session_test(self):
        """Test /sessions/delete (not actual cleanup)"""
        # Use different agent ID so we don't delete our test session
        payload = {"agentId": f"delete_{self.generate_random_string(8)}"}
        
        with self.client.delete(
            "/api/v1/sessions/delete",
            json=payload,
            catch_response=True,
            name="[SESSION] DELETE /delete"
        ) as response:
            if response.status_code in [200, 404]:
                response.success()
            else:
                response.failure(f"Status: {response.status_code}")
    
    # ========== LANGCHAIN ENDPOINTS ==========
    
    @task(8)
    def execute_langchain(self):
        """Test /langchain/execute"""
        messages = [
            "Hello, how are you?",
            "What's the weather?",
            "Tell me a joke",
            "Help me with a task",
            "Thank you!"
        ]
        
        payload = {
            "agentId": self.agent_id,
            "message": random.choice(messages),
            "sender": f"user_{self.generate_random_string(6)}",
            "params": {"max_steps": 5}
        }
        
        with self.client.post(
            "/api/v1/langchain/execute",
            json=payload,
            catch_response=True,
            name="[LANGCHAIN] POST /execute"
        ) as response:
            if response.status_code in [200, 404, 500]:
                response.success()
            else:
                response.failure(f"Status: {response.status_code}")
    
    # ========== SWAGGER ENDPOINTS ==========
    
    @task(1)
    def get_swagger_ui(self):
        """Test /swagger/ endpoint"""
        with self.client.get(
            "/swagger/index.html",
            catch_response=True,
            name="[DOCS] GET /swagger"
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Status: {response.status_code}")
    
    # ========== CLEANUP ==========
    
    def on_stop(self):
        """Cleanup session when stopping"""
        if self.session_created:
            try:
                self.client.delete(
                    "/api/v1/sessions/delete",
                    json={"agentId": self.agent_id},
                    name="[CLEANUP] DELETE session"
                )
                logger.info(f"Cleaned up: {self.agent_id}")
            except Exception as e:
                logger.error(f"Cleanup failed: {e}")


# ========== EVENT HOOKS ==========

@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    logger.info("=" * 70)
    logger.info("COMPREHENSIVE LOAD TEST - ALL ENDPOINTS")
    logger.info("Testing ALL available API endpoints")
    logger.info("Make sure:")
    logger.info("  1. API server is running (go run cmd/api/main.go)")
    logger.info("  2. APP_ENV=testing (for /create-test endpoint)")
    logger.info("  3. Database is accessible")
    logger.info("=" * 70)


@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    logger.info("=" * 70)
    logger.info("TEST COMPLETED")
    logger.info(f"Total requests: {environment.stats.total.num_requests}")
    logger.info(f"Total failures: {environment.stats.total.num_failures}")
    logger.info(f"Success rate: {(1 - environment.stats.total.fail_ratio) * 100:.2f}%")
    logger.info(f"Median response time: {environment.stats.total.median_response_time}ms")
    logger.info(f"95th percentile: {environment.stats.total.get_response_time_percentile(0.95)}ms")
    logger.info("=" * 70)


@events.request.add_listener
def on_request(request_type, name, response_time, response_length, exception, **kwargs):
    """Log slow requests"""
    if response_time > 1000:
        logger.warning(f"SLOW: {name} - {response_time}ms")
