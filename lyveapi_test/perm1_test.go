package lyveapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/steinfletcher/apitest"
	"github.com/szaydel/lyvecloud/lyveapi"
)

const mockPermission1 = "mock-permission-1"

const pCreateGoodRespBody = `{
	"id": "mock-permission-id"
  }`

func TestPermissions(t *testing.T) {
	const pGetByIdRespBody = `{
		"id": "mock-permission-1",
		"name": "mock-all",
		"description": "Mock description",
		"type": "bucket-names",
		"readyState": true,
		"actions": "all-operations",
		"prefix": "mock-prefix",
		"buckets": [
			"mock-bucket"
		],
		"policy": "mock-policy"
	}`
	const pListRespBody = `[{ 
		"name": "alpha",
		"id": "a1-b2-c3",
		"description": "something test", 
		"type": "all-buckets",
		"readyState": true,
		"createTime": "2023-08-11T19:15:00Z"
	   },
	   { 
		"name": "beta",
		"id": "c1-b2-a3",
		"description": "something test", 
		"type": "all-buckets",
		"readyState": true,
		"createTime": "2023-08-11T19:15:00Z"
	   }]`

	var permissionsListMock = apitest.NewMock().
		Get(mockPermissionsUri).
		RespondWith().
		Body(pListRespBody).
		Status(http.StatusOK).
		End()

	var permGetByIdMock = apitest.NewMock().
		Get(mockOnePermissionUri).
		RespondWith().
		Body(pGetByIdRespBody).
		Status(http.StatusOK).
		End()

	var permBadCreatePolicyMock = apitest.NewMock().
		Post(mockPermissionsUri).
		RespondWith().
		Body(createPermBadPolicyRespJSONObj).
		Status(http.StatusInternalServerError).
		End()

	var permGoodCreatePolicyMock = apitest.NewMock().
		Post(mockPermissionsUri).
		RespondWith().
		Body(pCreateGoodRespBody).
		Status(http.StatusOK).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(permissionsListMock).
		Handler(permissionsHandler()).
		Get(mockPermissionsUri).
		Expect(t).
		Body(pListRespBody).
		Status(http.StatusOK).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(permGetByIdMock).
		Handler(permissionsHandler()).
		Get(mockOnePermissionUri).
		Expect(t).
		Body(pGetByIdRespBody).
		Status(http.StatusOK).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(permBadCreatePolicyMock).
		Handler(permissionsHandler()).
		Post(mockPermissionsUri).
		Expect(t).
		Body(createPermBadPolicyRespJSONObj).
		Status(http.StatusInternalServerError).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(permGoodCreatePolicyMock).
		Handler(permsGoodCreateHandler()).
		Post(mockPermissionsUri).
		Expect(t).
		Body(pCreateGoodRespBody).
		Status(http.StatusOK).
		End()
}

func permissionsHandler() *http.ServeMux {
	var handler = http.NewServeMux()

	handler.HandleFunc(mockPermissionsUri, func(w http.ResponseWriter, r *http.Request) {
		var permission lyveapi.Permission
		var permissions lyveapi.PermissionList

		switch r.Method {
		case http.MethodGet:
			if err := permGet(mockPermissionsUri, &permissions); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			bytes, _ := json.Marshal(permissions)
			_, err := w.Write(bytes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)

		case http.MethodPost:
			if err := permsBadHttpPost(
				mockPermissionsUri, &permission); err != nil {
				// The order here is important. First, call the WriteHeader
				// method to set the http.StatusInternalServerError
				// response code. Next, write the necessary JSON payload.
				w.WriteHeader(
					err.(*lyveapi.ApiCallFailedError).HttpStatusCode())
				_, _ = w.Write([]byte(createPermBadPolicyRespJSONObj))
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	})

	handler.HandleFunc(mockOnePermissionUri, func(w http.ResponseWriter, r *http.Request) {
		var permission lyveapi.Permission
		if err := permGet(mockOnePermissionUri, &permission); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bytes, _ := json.Marshal(permission)
		_, err := w.Write(bytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	return handler
}

func permsGoodCreateHandler() *http.ServeMux {
	var handler = http.NewServeMux()

	handler.HandleFunc(mockPermissionsUri, func(w http.ResponseWriter, r *http.Request) {
		var permission lyveapi.Permission
		// var permissions lyveapi.PermissionList

		switch r.Method {
		case http.MethodPost:
			if err := permsGoodHttpPost(
				mockPermissionsUri, &permission); err != nil {
				// The order here is important. First, call the WriteHeader
				// method to set the http.StatusInternalServerError
				// response code. Next, write the necessary JSON payload.
				w.WriteHeader(
					err.(*lyveapi.ApiCallFailedError).HttpStatusCode())
				_, _ = w.Write([]byte(createPermBadPolicyRespJSONObj))
				return
			} else {
				w.WriteHeader(http.StatusOK)
				bytes, _ := json.Marshal(permission)
				_, err := w.Write(bytes)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

			}
		}
	})

	return handler
}

func permGet(path string, response interface{}) error {
	var err error
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockPermissionsUri:
		var p *lyveapi.PermissionList
		if p, err = client.ListPermissions(); err != nil {
			return err
		}
		pListPtr := response.(*lyveapi.PermissionList)
		*pListPtr = append(*pListPtr, *p...)
		return nil

	case mockOnePermissionUri:
		var p *lyveapi.Permission
		permissionId := mockOnePermissionUri[len(mockPermissionsUri+"/"):]
		if p, err = client.GetPermission(permissionId); err != nil {
			return err
		}
		response.(*lyveapi.Permission).Id = p.Id
		response.(*lyveapi.Permission).Name = p.Name
		response.(*lyveapi.Permission).Description = p.Description
		response.(*lyveapi.Permission).Type = p.Type
		response.(*lyveapi.Permission).ReadyState = p.ReadyState
		response.(*lyveapi.Permission).Actions = p.Actions
		response.(*lyveapi.Permission).Prefix = p.Prefix
		response.(*lyveapi.Permission).Buckets = p.Buckets
		response.(*lyveapi.Permission).Policy = p.Policy
		return nil
	}

	return pathUnmatchedErr
}

func permsGoodHttpPost(path string, response interface{}) error {
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	p := &lyveapi.Permission{
		Name:        "mock-permission-with-policy-1",
		Description: "Mock policy-type permission",
		Type:        lyveapi.Policy,
		Policy:      "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"statement15feb1\",\"Effect\":\"Allow\",\"Action\":[\"s3:*\"],\"Resource\":[\"arn:aws:s3:::*/*\"]}]}",
	}

	if resp, err := client.CreatePermission(p); err != nil {
		return err
	} else {
		response.(*lyveapi.Permission).Id = resp.Id
	}

	return nil
}

func permsBadHttpPost(path string, response interface{}) error {
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockPermissionsUri:
		policy := `{"garbage": "ploicy"}`
		permission := &lyveapi.Permission{
			Name:        mockPermission1,
			Description: "Mock description",
			Type:        lyveapi.Policy,
			Prefix:      "pre",
			Policy:      policy,
			Actions:     lyveapi.AllOperations}
		if _, err := client.CreatePermission(permission); err != nil {
			return err
		} else {
			return unexpectedSuccessErr
		}
	}

	return pathUnmatchedErr
}
