import random
from locust import task
from locust.contrib.fasthttp import FastHttpUser

class BoundedSearchUser(FastHttpUser):
    # No wait_time defined - users will make continuous requests
    
    search_terms = ["alpha", "beta", "electronics", "books", "product", "gamma", "delta"]
    
    @task
    def search_products(self):
        search_term = random.choice(self.search_terms)
        
        # Use name parameter to group all searches together
        self.client.get(f"/search?q={search_term}", name="/search")
