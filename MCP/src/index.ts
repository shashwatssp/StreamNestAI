interface Env {
	BACKEND_URL: string;
	MOVIE_CACHE: KVNamespace;
}

interface MCPRequest {
	jsonrpc: string;
	id: number | string;
	method: string;
	params?: any;
}

interface MCPResponse {
	jsonrpc: string;
	id: number | string;
	result?: any;
	error?: any;
}

export default {
	async fetch(request: Request, env: Env): Promise<Response> {
		const BACKEND_URL = env.BACKEND_URL;
		const url = new URL(request.url);


		const corsHeaders = {
			'Access-Control-Allow-Origin': '*',
			'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
			'Access-Control-Allow-Headers': 'Content-Type, Authorization',
		};

		if (request.method === 'OPTIONS') {
			return new Response(null, { headers: corsHeaders });
		}

		if (url.pathname === '/' || url.pathname === '') {
			if (request.method === 'GET') {
				return new Response('✅ StreamNestAI MCP Server is running! Use POST / for MCP requests', {
					headers: { ...corsHeaders, 'Content-Type': 'text/plain' },
				});
			}

			if (request.method === 'POST') {
				try {
					const mcpRequest: MCPRequest = await request.json();

					const response = await handleMCPRequest(mcpRequest, BACKEND_URL, env);


					return new Response(JSON.stringify(response), {
						headers: {
							...corsHeaders,
							'Content-Type': 'application/json',
						},
					});
				} catch (error: any) {
					console.error('Parse error:', error.message);
					return new Response(JSON.stringify({
						jsonrpc: '2.0',
						id: null,
						error: { code: -32700, message: 'Parse error', data: error.message },
					}), {
						status: 400,
						headers: { ...corsHeaders, 'Content-Type': 'application/json' },
					});
				}
			}
		}

		return new Response('Method not allowed', { status: 405, headers: corsHeaders });
	},
};

async function handleMCPRequest(request: MCPRequest, BACKEND_URL: string, env: Env): Promise<MCPResponse> {

	const { method, params, id } = request;


	try {
		if (method === 'initialize') {
			return {
				jsonrpc: '2.0',
				id,
				result: {
					protocolVersion: '2024-11-05',
					capabilities: { tools: {} },
					serverInfo: { name: 'StreamNestAI', version: '1.0.0' },
				},
			};
		}

		if (method === 'tools/list') {
			return {
				jsonrpc: '2.0',
				id,
				result: {
					tools: [
						{
							name: 'search_all_movies',
							description: 'Get all movies in the database',
							inputSchema: { type: 'object', properties: {} },
						},
						{
							name: 'get_movie_by_id',
							description: 'Get movie details by IMDb ID',
							inputSchema: {
								type: 'object',
								properties: { imdb_id: { type: 'string' } },
								required: ['imdb_id'],
							},
						},
						{
							name: 'get_recommended_movies',
							description: 'Get top-rated movies',
							inputSchema: { type: 'object', properties: {} },
						},
						{
							name: 'get_genres',
							description: 'Get all genres',
							inputSchema: { type: 'object', properties: {} },
						},
						{
							name: 'search_by_keyword',
							description: 'Search movies by keyword',
							inputSchema: {
								type: 'object',
								properties: { keyword: { type: 'string' } },
								required: ['keyword'],
							},
						},
					],
				},
			};
		}

		if (method === 'tools/call') {
			const { name, arguments: args } = params;

			const toolResult = await executeToolCall(name, args, BACKEND_URL, env.MOVIE_CACHE);


			return {
				jsonrpc: '2.0',
				id,
				result: {
					content: [{ type: 'text', text: JSON.stringify(toolResult, null, 2) }],
				},
			};
		}

		return {
			jsonrpc: '2.0',
			id,
			error: { code: -32601, message: `Method not found: ${method}` },
		};
	} catch (error: any) {
		console.error('Error in handleMCPRequest:', error.message);
		console.error('Stack:', error.stack);
		return {
			jsonrpc: '2.0',
			id,
			error: { code: -32603, message: 'Internal error', data: error.message },
		};
	}
}

async function executeToolCall(toolName: string, args: any, BACKEND_URL: string, CACHE: KVNamespace): Promise<any> {

	switch (toolName) {
		case 'search_all_movies': {
			const cacheKey = 'all_movies';

			// Try cache first
			const cached = await CACHE.get(cacheKey, 'json');
			if (cached) {
				console.log('✅ Cache HIT - all_movies');
				return cached;
			}

			// Cache miss - fetch from backend
			console.log('❌ Cache MISS - fetching all_movies');
			const url = `${BACKEND_URL}/movies`;
			const res = await fetch(url);
			const text = await res.text();
			const data = JSON.parse(text);

			if (!Array.isArray(data)) {
				throw new Error(`Not an array`);
			}

			// Store in cache for 10 hours
			await CACHE.put(cacheKey, JSON.stringify(data), {
				expirationTtl: 36000,
			});

			console.log('✅ Cached all_movies for 1 hour');
			return data;
		}


		case 'get_movie_by_id': {
			const { imdb_id } = args;
			const cacheKey = `movie_${imdb_id}`;

			// Try cache first
			const cached = await CACHE.get(cacheKey, 'json');
			if (cached) {
				console.log('✅ Cache HIT - movie:', imdb_id);
				return cached;
			}

			// Cache miss
			console.log('❌ Cache MISS - fetching movie:', imdb_id);
			const url = `${BACKEND_URL}/movies/${imdb_id}`;
			const res = await fetch(url);
			const data = await res.json();

			if (data.error || !res.ok) {
				throw new Error('Movie not found');
			}

			// Store in cache for 24 hours
			await CACHE.put(cacheKey, JSON.stringify(data), {
				expirationTtl: 86400,
			});

			console.log('✅ Cached movie:', imdb_id);
			return data;
		}


		case 'get_recommended_movies': {
			const url = `${BACKEND_URL}/movies`;

			const res = await fetch(url);

			const text = await res.text();

			let data;
			try {
				data = JSON.parse(text);
			} catch (e) {
				console.error('❌ JSON parse error:', e);
				throw new Error(`Invalid JSON: ${text.substring(0, 100)}`);
			}

			if (data.error) {
				throw new Error(data.error);
			}

			if (!Array.isArray(data)) {
				console.error('❌ Not an array, got:', typeof data);
				throw new Error(`Backend returned non-array for movies: ${JSON.stringify(data).substring(0, 100)}`);
			}


			const topRated = data
				.filter((m: any) => {
					const hasRanking = m.ranking && typeof m.ranking.ranking_value === 'number';
					return hasRanking && m.ranking.ranking_value <= 3;
				})
				.sort((a: any, b: any) => {
					const rankA = a.ranking?.ranking_value || 999;
					const rankB = b.ranking?.ranking_value || 999;
					return rankA - rankB;
				})
				.slice(0, 10);


			if (topRated.length === 0) {
				return data.slice(0, 10);
			}

			return topRated;
		}

		case 'get_genres': {
			const cacheKey = 'genres';

			// Try cache first
			const cached = await CACHE.get(cacheKey, 'json');
			if (cached) {
				console.log('✅ Cache HIT - genres');
				return cached;
			}

			// Cache miss
			console.log('❌ Cache MISS - fetching genres');
			const url = `${BACKEND_URL}/genres`;
			const res = await fetch(url);
			const data = await res.json();

			if (!Array.isArray(data)) {
				throw new Error(`Not an array`);
			}

			// Store in cache for 24 hours
			await CACHE.put(cacheKey, JSON.stringify(data), {
				expirationTtl: 86400,
			});

			console.log('✅ Cached genres for 24 hours');
			return data;
		}


case 'search_by_keyword': {
    const { keyword } = args;
    const cacheKey = `search_${keyword.toLowerCase()}`;

    // Try cache first
    const cached = await CACHE.get(cacheKey, 'json');
    if (cached) {
        console.log('✅ Cache HIT - search:', keyword);
        return cached;
    }

    // Cache miss
    console.log('❌ Cache MISS - searching:', keyword);
    const url = `${BACKEND_URL}/movies`;
    console.log('Fetching from:', url);  // ← ADD THIS
    
    const res = await fetch(url);
    console.log('Response status:', res.status);  // ← ADD THIS
    
    if (!res.ok) {
        throw new Error(`Backend returned ${res.status}: ${res.statusText}`);
    }
    
    const text = await res.text();
    console.log('Response text length:', text.length);  // ← ADD THIS
    console.log('First 100 chars:', text.substring(0, 100));  // ← ADD THIS
    
    // Check if response is empty
    if (!text || text.trim() === '') {
        throw new Error('Backend returned empty response');
    }
    
    let data;
    try {
        data = JSON.parse(text);
    } catch (e) {
        console.error('Failed to parse JSON:', text.substring(0, 200));
        throw new Error(`Invalid JSON from backend: ${text.substring(0, 100)}`);
    }

    if (!Array.isArray(data)) {
        console.error('Response is not an array:', typeof data);
        throw new Error(`Backend returned non-array: ${JSON.stringify(data).substring(0, 100)}`);
    }

    console.log('Got', data.length, 'movies from backend');

    const keywordLower = keyword.toLowerCase();
    const filtered = data.filter((m: any) => {
        const titleMatch = m.title && m.title.toLowerCase().includes(keywordLower);
        const genreMatch = m.genre && Array.isArray(m.genre) && 
            m.genre.some((g: any) => g.genre_name && g.genre_name.toLowerCase().includes(keywordLower));
        return titleMatch || genreMatch;
    });

    // Store search results for 100 minutes
    await CACHE.put(cacheKey, JSON.stringify(filtered), {
        expirationTtl: 6000,
    });

    console.log('✅ Cached search:', keyword, '- found:', filtered.length);
    return filtered;
}




		default:
			throw new Error(`Unknown tool: ${toolName}`);
	}
}
