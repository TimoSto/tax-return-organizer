// Response shapes of the backend REST API. Keep in sync with backend/api/openapi.yaml.

/** GET /auth/session */
export interface GetSessionResponse {
	role: 'staff' | 'client';
	userId: string;
	taxPracticeId: string;
}
