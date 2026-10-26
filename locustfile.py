import os
import random

from locust import FastHttpUser, LoadTestShape, constant, task

TERMS = ["electronics", "Electro", "books", "HOME", "garden", "toys",
         "alpha", "Beta", "nova", "product", "zzz-nomatch"]


class SearchUser(FastHttpUser):
    wait_time = constant(0)  # minimal wait: next request as soon as the last returns

    @task
    def search(self):
        q = random.choice(TERMS)
        with self.client.get(f"/products/search?q={q}", name="/products/search",
                             catch_response=True) as r:
            if r.status_code != 200:
                r.failure(f"HTTP {r.status_code}")
            elif "total_found" not in r.text:
                r.failure("missing total_found")


SHAPE = os.getenv("SHAPE")

if SHAPE == "ramp":
    class GradualRamp(LoadTestShape):
        """+5 users every minute from 5 to 40, stop at 15 min."""
        def tick(self):
            t = self.get_run_time()
            if t > 900:
                return None
            return min(40, 5 + 5 * int(t // 60)), 5

elif SHAPE == "spike":
    class SuddenSpike(LoadTestShape):
        """3 min quiet, 5 min spike, 5 min quiet."""
        stages = [(180, 5), (480, 40), (780, 5)]

        def tick(self):
            t = self.get_run_time()
            for end, users in self.stages:
                if t < end:
                    return users, 40
            return None
