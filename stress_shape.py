from locust import LoadTestShape


class StepShape(LoadTestShape):
    # (stage ends at seconds, users, spawn rate)
    stages = [
        (120, 50, 10),
        (240, 200, 20),
        (420, 500, 50),
        (600, 1000, 50),
    ]

    def tick(self):
        t = self.get_run_time()
        for end, users, rate in self.stages:
            if t < end:
                return users, rate
        return None  # stop the test
