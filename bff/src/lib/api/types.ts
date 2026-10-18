// Response shapes of the backend REST API. Keep in sync with backend/api/openapi.yaml.

/** GET /collectors/{id} */
export interface GetCollectorResponse {
	id: string;
	name: string;
	email: string;
}
