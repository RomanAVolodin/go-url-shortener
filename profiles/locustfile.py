import random

from locust import HttpUser, task


class HelloWorldUser(HttpUser):
    @task
    def delete_urls(self):
        self.client.delete(
            '/api/user/urls', json=['EX7PoGHwZpPpusdhiWFm5F', 'XMpN2csaRdN2V2bEnyjcB8', 'xvuHQQ9po9KbcG65Uo2zSP']
        )

    @task
    def create_urls_by_json(self):
        x = random.randint(1, 100000)
        self.client.post('/api/shorten', json={'url': f'https://many_{x}.ru'})

    @task
    def create_urls_by_json(self):
        x = random.randint(1, 100000)
        self.client.post('/', data=f'https://many_{x}.ru')

    @task
    def get_users_records(self):
        self.client.get('/api/user/urls',)

    @task
    def create_multiple(self):
        x = random.randint(1, 100000)
        self.client.post(
            '/api/shorten/batch',
            json=[
                {'correlation_id': '97892036-cb8e-45ca-bc22-add747c970ef', 'original_url': f'http://xepnobhj{x}.biz'},
                {'correlation_id': 'bd2fc952-7d60-4351-ab8f-8a2f878aa05a', 'original_url': f'http://rra{x}.yandex'},
            ],
        )
