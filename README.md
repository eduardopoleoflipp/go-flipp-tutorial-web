curl commands to use this

To see posts

```bash
curl http://localhost:3000/posts/index | jq
```
you can instal jq by `brew install jq`


To Create a post
```bash
curl -X POST http://localhost:3000/posts/create \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Post",
    "content": "This is the content of the post.",
    "created_at": "2025-04-30T12:00:00Z",
    "author": "Jane Doe"
  }'

```