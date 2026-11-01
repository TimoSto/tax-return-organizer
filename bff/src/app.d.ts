// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
import type { GetSessionResponse } from '#lib/api/types.js';

declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			/** The caller, as the backend resolved its session; null without a valid session. Set by handleAuth. */
			actor: GetSessionResponse | null;
		}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
