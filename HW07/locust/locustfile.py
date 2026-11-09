import random
from locust import HttpUser, task, between


def make_order():
    """Build a random order body."""
    return {
        "customer_id": random.randint(1, 10000),
        "items": [{"product_id": "p{}".format(random.randint(1, 50)), "quantity": random.randint(1, 3)}],
    }


class SyncUser(HttpUser):
    """Customer who waits for payment to finish."""
    wait_time = between(0.1, 0.5)

    @task
    def place_order(self):
        self.client.post("/orders/sync", json=make_order(), name="POST /orders/sync")


class AsyncUser(HttpUser):
    """Customer who gets an instant 202."""
    wait_time = between(0.1, 0.5)

    @task
    def place_order(self):
        self.client.post("/orders/async", json=make_order(), name="POST /orders/async")
