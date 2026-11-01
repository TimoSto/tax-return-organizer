// Response shapes of the backend REST API. Keep in sync with backend/api/openapi.yaml.

/** GET /tax-practice */
export interface GetTaxPracticeResponse {
	id: string;
	name: string;
	email: string;
}

/** GET /auth/session */
export interface GetSessionResponse {
	role: 'staff' | 'client';
	userId: string;
	taxPracticeId: string;
}
