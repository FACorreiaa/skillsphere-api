{
"mappings": {
"properties": {
"user_id": { "type": "keyword" },
"username": { "type": "keyword" },
"display_name": { "type": "text" },
"bio": { "type": "text" },

      "skills_offered": {
        "type": "nested",
        "properties": {
          "id": { "type": "keyword" },
          "name": { "type": "text" },
          "category": { "type": "keyword" },
          "proficiency": { "type": "integer" }
        }
      },

      "skills_wanted": {
        "type": "nested",
        "properties": {
          "id": { "type": "keyword" },
          "name": { "type": "text" },
          "category": { "type": "keyword" },
          "proficiency": { "type": "integer" }
        }
      },

      "location": { "type": "geo_point" },
      "availability": { "type": "keyword" },

      "average_rating": { "type": "float" },
      "total_sessions": { "type": "integer" },
      "is_verified": { "type": "boolean" },

      "created_at": { "type": "date" }
    }
}
}
