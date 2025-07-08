
#Search near a point (50km radius):
curl -X POST http://localhost:8080/search/advanced \
  -H "Content-Type: application/json" \
  -d '{
    "nearPoint": [40.7128, -74.0060],  # New York coordinates
    "withinRadius": 50,                # 50km radius
    "limit": 10
  }'

#Search within a bounding box:
curl -X POST http://localhost:8080/search/advanced \
  -H "Content-Type: application/json" \
  -d '{
    "boundingBox": [40.7, -74.1, 40.8, -73.9],  # minLat, minLon, maxLat, maxLon
    "mediaType": "image",
    "limit": 20
  }'