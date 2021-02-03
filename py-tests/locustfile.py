import time
from locust import HttpUser, task, between


class QuickstartUser(HttpUser):
    wait_time = between(1, 2.5)

    @task
    def hello_world(self):
        headers = {'Authorization':'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MTIzNDc3OTcsImlhdCI6MTYxMjM0NDE5NywidXNlciI6Imthc2ltb3YzMjExQGdtYWlsLmNvbSJ9.FAEeda9Pw0_e-RW4p1UqD87QLTFXjrIi6WQ_KWINcFc'}
        self.client.get("/logs/Rostov-On-Don", headers=headers)

    @task
    def view_items(self):
        self.client.get("/token?email=kasimov3211@gmail.com&password=12345")
        time.sleep(1)

    def on_start(self):
        headers = {'Authorization':'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MTIzNDc3OTcsImlhdCI6MTYxMjM0NDE5NywidXNlciI6Imthc2ltb3YzMjExQGdtYWlsLmNvbSJ9.FAEeda9Pw0_e-RW4p1UqD87QLTFXjrIi6WQ_KWINcFc'}
        self.client.get("/weather/Rostov-On-Don", headers=headers)
