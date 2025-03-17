# edduhub


Backend kratos 

base) tgt@tgt:~/Desktop/edduhub$ curl -s -X GET \
  -H "Accept: application/json" \
  "http://localhost:4433/self-service/registration/api" | jq
{
  "id": "3bd746e4-50c2-4dea-ac7a-3ebe4fc899f2",
  "oauth2_login_challenge": null,
  "type": "api",
  "expires_at": "2025-03-17T07:15:52.409950388Z",
  "issued_at": "2025-03-17T07:05:52.409950388Z",
  "request_url": "http://localhost:4433/self-service/registration/api",
  "ui": {
    "action": "http://127.0.0.1:4433/self-service/registration?flow=3bd746e4-50c2-4dea-ac7a-3ebe4fc899f2",
    "method": "POST",
    "nodes": [
      {
        "type": "input",
        "group": "default",
        "attributes": {
          "name": "csrf_token",
          "type": "hidden",
          "value": "",
          "required": true,
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {}
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "traits.email",
          "type": "email",
          "required": true,
          "autocomplete": "email",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070002,
            "text": "E-Mail",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "password",
          "type": "password",
          "required": true,
          "autocomplete": "new-password",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070001,
            "text": "Password",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "traits.name.first",
          "type": "text",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070002,
            "text": "First Name",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "traits.name.last",
          "type": "text",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070002,
            "text": "Last Name",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "traits.college.id",
          "type": "text",
          "required": true,
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070002,
            "text": "College ID",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "traits.college.name",
          "type": "text",
          "required": true,
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070002,
            "text": "College Name",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "traits.role",
          "type": "text",
          "required": true,
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070002,
            "text": "Role",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "method",
          "type": "submit",
          "value": "password",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1040001,
            "text": "Sign up",
            "type": "info",
            "context": {}
          }
        }
      }
    ]
  }
}
(base) tgt@tgt:~/Desktop/edduhub$ curl -s -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "method": "password",
    "password": "secure-password123",
    "traits": {
      "email": "test@example.com",
      "name": {
        "first": "Test",
        "last": "User"
      },
      "college": {
        "id": "college123",
        "name": "Test College"
      },
      "role": "student"
    }
  }' \
  "http://localhost:4433/self-service/registration?flow=1" | jq
{
  "error": {
    "code": 400,
    "status": "Bad Request",
    "reason": "The flow query parameter is missing or malformed.",
    "message": "The request was malformed or contained invalid parameters"
  }
}
(base) tgt@tgt:~/Desktop/edduhub$ curl -s -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "method": "password",
    "password": "secure-password123",
    "traits": {
      "email": "test@example.com",
      "name": {
        "first": "Test",
        "last": "User"
      },
      "college": {
        "id": "college123",
        "name": "Test College"
      },
      "role": "student"
    }
  }' \
  "http://localhost:4433/self-service/registration?flow=3bd746e4-50c2-4dea-ac7a-3ebe4fc899f2" | jq
{
  "session_token": "ory_st_TNiTbefZF4qNkWr9DGIccMp8upaxzBZc",
  "session": {
    "id": "7efedf89-4dda-4636-8320-7506e58db1a3",
    "active": true,
    "expires_at": "2025-03-18T07:08:12.804256949Z",
    "authenticated_at": "2025-03-17T07:08:12.806935832Z",
    "authenticator_assurance_level": "aal1",
    "authentication_methods": [
      {
        "method": "password",
        "aal": "aal1",
        "completed_at": "2025-03-17T07:08:12.804312476Z"
      }
    ],
    "issued_at": "2025-03-17T07:08:12.804256949Z",
    "identity": {
      "id": "6a6c4984-f03c-4539-b9fa-3f6d23c4faf6",
      "schema_id": "default",
      "schema_url": "http://127.0.0.1:4433/schemas/ZGVmYXVsdA",
      "state": "active",
      "state_changed_at": "2025-03-17T07:08:12.802576576Z",
      "traits": {
        "email": "test@example.com",
        "name": {
          "first": "Test",
          "last": "User"
        },
        "college": {
          "id": "college123",
          "name": "Test College"
        },
        "role": "student"
      },
      "verifiable_addresses": [
        {
          "id": "34c311b1-1452-4c00-9094-253a25b1cdc6",
          "value": "test@example.com",
          "verified": false,
          "via": "email",
          "status": "sent",
          "created_at": "2025-03-17T07:08:12.803594Z",
          "updated_at": "2025-03-17T07:08:12.803594Z"
        }
      ],
      "recovery_addresses": [
        {
          "id": "27b2f19d-a918-416f-9f1d-1e231cccb4bd",
          "value": "test@example.com",
          "via": "email",
          "created_at": "2025-03-17T07:08:12.803735Z",
          "updated_at": "2025-03-17T07:08:12.803735Z"
        }
      ],
      "metadata_public": null,
      "created_at": "2025-03-17T07:08:12.803393Z",
      "updated_at": "2025-03-17T07:08:12.803393Z"
    },
    "devices": [
      {
        "id": "e564cf4f-fa03-4e39-a4ba-ac0ae93ae581",
        "ip_address": "172.20.0.1:65332",
        "user_agent": "curl/8.5.0",
        "location": ""
      }
    ]
  },
  "identity": {
    "id": "6a6c4984-f03c-4539-b9fa-3f6d23c4faf6",
    "schema_id": "default",
    "schema_url": "http://127.0.0.1:4433/schemas/ZGVmYXVsdA",
    "state": "active",
    "state_changed_at": "2025-03-17T07:08:12.802576576Z",
    "traits": {
      "email": "test@example.com",
      "name": {
        "first": "Test",
        "last": "User"
      },
      "college": {
        "id": "college123",
        "name": "Test College"
      },
      "role": "student"
    },
    "verifiable_addresses": [
      {
        "id": "34c311b1-1452-4c00-9094-253a25b1cdc6",
        "value": "test@example.com",
        "verified": false,
        "via": "email",
        "status": "sent",
        "created_at": "2025-03-17T07:08:12.803594Z",
        "updated_at": "2025-03-17T07:08:12.803594Z"
      }
    ],
    "recovery_addresses": [
      {
        "id": "27b2f19d-a918-416f-9f1d-1e231cccb4bd",
        "value": "test@example.com",
        "via": "email",
        "created_at": "2025-03-17T07:08:12.803735Z",
        "updated_at": "2025-03-17T07:08:12.803735Z"
      }
    ],
    "metadata_public": null,
    "created_at": "2025-03-17T07:08:12.803393Z",
    "updated_at": "2025-03-17T07:08:12.803393Z"
  },
  "continue_with": [
    {
      "action": "show_verification_ui",
      "flow": {
        "id": "ee378b7b-2c3e-4d5e-be9a-9441fb502b8c",
        "verifiable_address": "test@example.com"
      }
    },
    {
      "action": "set_ory_session_token",
      "ory_session_token": "ory_st_TNiTbefZF4qNkWr9DGIccMp8upaxzBZc"
    }
  ]
}