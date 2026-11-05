// Response shapes of the backend REST API. Keep in sync with backend/api/openapi.yaml.

export interface ValueSpec {
	type: 'text' | 'number' | 'date' | 'bool';
	/** Free-text display label such as EUR; empty if there is none. */
	unit: string;
	required: boolean;
}

export interface MetadataField {
	name: string;
	valueSpec: ValueSpec;
}

export interface DocumentSpec {
	metadataFields: MetadataField[];
}

/** Something a persona asks the client to provide. A document item has a documentSpec, a data item a valueSpec. */
export interface ItemDefinition {
	id: string;
	title: string;
	kind: 'document' | 'data';
	recurrence: 'monthly' | 'quarterly' | 'semiannually' | 'annually';
	documentSpec: DocumentSpec | null;
	valueSpec: ValueSpec | null;
	/** Empty for a single item; otherwise the item is repeatable and the first field labels the instance. */
	instanceFields: MetadataField[];
}

/** GET /personas/{id} */
export interface GetPersonaResponse {
	id: string;
	taxPracticeId: string;
	name: string;
	/** May be empty. */
	description: string;
	/** In the order the tax practice arranged them. */
	items: ItemDefinition[];
}

/** GET /personas */
export interface ListPersonasResponse {
	/** In no particular order. */
	personas: GetPersonaResponse[];
}
