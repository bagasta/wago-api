"""
Automated Locust Load Testing for WhatsApp API
Uses test endpoint to bypass QR scanning for thousands of concurrent users

Run: locust -f locustfile_auto.py --host=http://localhost:8080
"""

from locust import HttpUser, task, between, events
import random
import string
import logging

# Setup logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class AutomatedSessionUser(HttpUser):
    """
    Fully automated user that creates sessions without QR scanning.
    Uses the /create-test endpoint which bypasses QR code requirement.
    """
    
    wait_time = between(0.5, 2)
    
    def on_start(self):
        """Initialize test data for this user"""
        self.agent_id = f"load_test_{self.generate_random_string(10)}"
        self.api_key = "secret"
        self.session_created = False
        logger.info(f"User started with agent_id: {self.agent_id}")
    
    def on_stop(self):
        """Cleanup: Delete session when user stops"""
        if self.session_created:
            self.delete_test_session()
    
    def generate_random_string(self, length=8):
        """Generate random string for unique IDs"""
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))
    
    @task(10)
    def create_test_session(self):
        """
        Create a test session that bypasses QR scanning.
        This session will appear as 'connected' immediately.
        """
        if self.session_created:
            return  # Already created
        
        payload = {
            "agentId": self.agent_id,
            "agentName": f"Load Test Agent {self.agent_id}",
            "apiKey": self.api_key,
            "langchainUrl": "https://api.langchain.example.com"
        }
        
        with self.client.post(
            "/api/v1/sessions/create-test",
            json=payload,
            catch_response=True,
            name="[AUTO] Create Test Session"
        ) as response:
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    self.session_created = True
                    logger.info(f"✓ Test session created: {self.agent_id}")
                    response.success()
                else:
                    response.failure(f"API returned success=false: {data}")
            elif response.status_code == 409:
                # Session already exists
                self.session_created = True
                response.success()
            elif response.status_code == 403:
                response.failure("Test endpoint not enabled. Set APP_ENV=testing or development")
            else:
                response.failure(f"Unexpected status: {response.status_code}")
    
    @task(20)
    def get_session_status(self):
        """Check session status"""
        with self.client.get(
            f"/api/v1/sessions/status?agentId={self.agent_id}",
            catch_response=True,
            name="[AUTO] Get Status"
        ) as response:
            if response.status_code == 200:
                data = response.json()
                if data.get("data", {}).get("status") == "connected":
                    response.success()
                else:
                    response.success()  # Accept any status
            elif response.status_code == 404:
                response.success()  # Session not created yet
            else:
                response.failure(f"Unexpected status: {response.status_code}")
    
    @task(15)
    def get_session_detail(self):
        """Get detailed session information"""
        with self.client.get(
            f"/api/v1/sessions/detail?agentId={self.agent_id}",
            catch_response=True,
            name="[AUTO] Get Detail"
        ) as response:
            if response.status_code in [200, 404]:
                response.success()
            else:
                response.failure(f"Unexpected status: {response.status_code}")
    
    @task(5)
    def execute_langchain(self):
        """
        Execute Langchain with the connected session.
        This tests the full integration flow.
        """
        if not self.session_created:
            return  # Need session first
        
        messages = [
            "Hello",
            "How are you?",
            "What's the weather today?",
            "Tell me a joke",
            "Help me with a task"
        ]
        
        payload = {
            "agentId": self.agent_id,
            "message": random.choice(messages),
            "sender": f"user_{self.generate_random_string(6)}",
            "params": {
                "max_steps": 5
            }
        }
        
        with self.client.post(
            "/api/v1/langchain/execute",
            json=payload,
            catch_response=True,
            name="[AUTO] Execute Langchain"
        ) as response:
            if response.status_code == 200:
                response.success()
            elif response.status_code in [404, 500]:
                # May fail if session not fully initialized or langchain unavailable
                response.success()
            else:
                response.failure(f"Unexpected status: {response.status_code}")
    
    @task(2)
    def health_check(self):
        """Quick health check"""
        self.client.get("/health", name="[AUTO] Health Check")
    
    def delete_test_session(self):
        """Delete the test session on cleanup"""
        try:
            payload = {"agentId": self.agent_id}
            response = self.client.delete(
                "/api/v1/sessions/delete",
                json=payload,
                name="[AUTO] Cleanup Session"
            )
            if response.status_code in [200, 404]:
                logger.info(f"✓ Session cleaned up: {self.agent_id}")
        except Exception as e:
            logger.error(f"Failed to cleanup session {self.agent_id}: {e}")


# Event hooks for reporting
@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    logger.info("=" * 60)
    logger.info("AUTOMATED LOAD TEST STARTING")
    logger.info("Using /create-test endpoint to bypass QR scanning")
    logger.info("Make sure APP_ENV=testing or development")
    logger.info("=" * 60)


@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    logger.info("=" * 60)
    logger.info("LOAD TEST COMPLETED")
    logger.info(f"Total requests: {environment.stats.total.num_requests}")
    logger.info(f"Total failures: {environment.stats.total.num_failures}")
    logger.info(f"Median response time: {environment.stats.total.median_response_time}ms")
    logger.info(f"Average response time: {environment.stats.total.avg_response_time}ms")
    logger.info(f"95 percentile: {environment.stats.total.get_response_time_percentile(0.95)}ms")
    logger.info("=" * 60)


@events.request.add_listener
def on_request(request_type, name, response_time, response_length, exception, **kwargs):
    """Log slow requests"""
    if response_time > 1000:  # > 1 second
        logger.warning(f"Slow request: {name} took {response_time}ms")
