# UHC Auth Proxy

uhc-auth-proxy is a service for authenticating Openshift 4 clusters. It does
so by handling the various API calls required to authenticate a `cluster_id`
and `authorization_token` with UHC services.

The `cluster_id` and `authorization_token` are two pieces of data that are
stored with each deployment of an Openshift 4 cluster. They are issued when
the cluster is provisioned.

This service provides an alternative way to authenticate and authorize
requests without requiring a customer to store their SSO credentials
somewhere within their cluster. This, in turn, enables the support-operator
to send data with no additional configuration required.

## How it works

1. A request is made containing the cluster_id and authorization_token from a
   cluster
2. the auth proxy requests the account_id that provisioned the cluster. Here
   account_id refers to a single user.
3. The auth proxy requests the organzation_id that the provisioning user is a
   member of.
4. The auth proxy requests organization details for the found
   organization_id.
5. Assuming everything above works, an identification document is returned
   that looks something like this:

    {
        "account_number": "123456",
        "type": "system",
        "internal": {
            "org_id": "1234567"
        }
    }

## Integration with current API broker

In practice, a request is made of the API broker from the support-operator
with the following headers:

    User-agent: support-operator/<git hash> cluster/<cluster id>
    Authentication: Bearer <authorization_token>

Because the header indicates the call came from the support-operator the
uhc-auth-proxy is delegated to for authentication.

## Configuration

In order to communicate with the backend services that UHC uses a priveleged
service account must be maintained. This account will have an
`offline_access_token` that should be configured as part of the deployment of
the service.

The _oat_ can be used to request a short-lived token that will be allowed to
access the UHC services.

The following APIs will need to be accessed:

### POST https://api.openshift.com/api/accounts_mgmt/v1/cluster_registrations

A document like below is posted:

    {
        "cluster_id": "$cluster_id",
        "authorization_token": "$authorization_token",
    }

The API returns something like:

    {
        "cluster_id": "$cluster_id",
        "authorization_token": "$authorization_token",
        "account_id": "string",
        "expires_at": "number"
    }

### GET https://api.openshift.com/api/accounts_mgmt/v1/accounts/:account_id

The API returns something like:

    {
        "id": "string",
        "kind": "Account",
        "href": "/api/accounts_mgmt/v1/accounts/:account_id",
        "first_name": "string",
        "last_name": "string",
        "username": "string",
        "email": "string",
        "banned": false,
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "organization": {
            "id": "string",
            "kind": "Organization",
            "href": "/api/accounts_mgmt/v1/organizations/:org_id",
            "name": "Organization Name"
        }
    }

### GET https://api.openshift.com/api/accounts_mgmt/v1/organizations/:org_id

The API returns something like:

    {
        "id": ":org_id",
        "kind": "Organization",
        "href": "/api/accounts_mgmt/v1/organizations/:org_id",
        "name": "Organization Name",
        "external_id": "number",
        "ebs_account_id": "number",
        "created_at": "timestamp",
        "updated_at": "timestamp"
    }

Permissions must be granted and maintained for each resource.

## CLI

You can try all the calls via the cli. To get started run `go install` from
the root of the project.

This will build and install the `uhc-auth-proxy` command. You can request an
identity document like this:

    uhc-auth-proxy --oat $OFFLINE_AUTH_TOKEN --cluster-id $CLUSTER_ID --authorization-token $AUTHORIZATION_TOKEN
