import os
import random

from locust import FastHttpUser, HttpUser, between, constant, task

BaseUser = HttpUser if os.getenv("LOCUST_CLIENT", "fast").lower() == "http" else FastHttpUser

MAX_ID = 1000
GET_WEIGHT = int(os.getenv("GET_WEIGHT", "9"))
POST_WEIGHT = int(os.getenv("POST_WEIGHT", "1"))
BAD_WEIGHT = int(os.getenv("BAD_WEIGHT", "0"))


class ProductUser(BaseUser):
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

    @task(BAD_WEIGHT)
    def get_missing(self):
        self.client.get("/products/999999", name="/products/[missing]")
