import os
import random

from locust import FastHttpUser, HttpUser, between, constant, task

# HW3 used FastHttpUser. Set LOCUST_CLIENT=http to run the same test with HttpUser.
BaseUser = HttpUser if os.getenv("LOCUST_CLIENT", "fast").lower() == "http" else FastHttpUser

MAX_ID = 1000  # products 1..1000 are seeded at startup
GET_WEIGHT = int(os.getenv("GET_WEIGHT", "9"))
POST_WEIGHT = int(os.getenv("POST_WEIGHT", "1"))


class ProductUser(BaseUser):
    # HW3 had no wait time (maximum pressure). THINK=1 adds 0.1-0.5 s between requests.
    wait_time = between(0.1, 0.5) if os.getenv("THINK") == "1" else constant(0)

    @task(GET_WEIGHT)
    def get_product(self):
        self.client.get(f"/products/{random.randint(1, MAX_ID)}", name="/products/[id]")

    @task(POST_WEIGHT)
    def update_details(self):
        pid = random.randint(1, MAX_ID)
        self.client.post(
            f"/products/{pid}/details",
            json={
                "product_id": pid,
                "sku": f"SKU-{pid}-{random.randint(1, 9999)}",
                "manufacturer": "Acme",
                "category_id": 1,
                "weight": 10,
                "some_other_id": 1,
            },
            name="/products/[id]/details",
        )
