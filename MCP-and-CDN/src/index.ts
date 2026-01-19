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

// Comprehensive logging utility
const logMCPFeature = (action: string, data: any = {}, level: string = 'info') => {
	const timestamp = new Date().toISOString();
	const logEntry = {
		timestamp,
		feature: 'MCP_CDN',
		action,
		data,
		level
	};
	
	const logMessage = `[${timestamp}] 🌐 [MCP_CDN] ${action}: ${JSON.stringify(data)}`;
	
	switch(level) {
		case 'error':
			console.error(logMessage, logEntry);
			break;
		case 'warn':
			console.warn(logMessage, logEntry);
			break;
		case 'debug':
			console.debug(logMessage, logEntry);
			break;
		default:
			console.log(logMessage, logEntry);
	}
};

export default {
	async fetch(request: Request, env: Env): Promise<Response> {
		const startTime = Date.now();
		const BACKEND_URL = env.BACKEND_URL;
		const url = new URL(request.url);
		const requestId = Math.random().toString(36).substring(7);

		logMCPFeature('REQUEST_START', {
			requestId,
			method: request.method,
			url: url.pathname,
			userAgent: request.headers.get('user-agent'),
			origin: request.headers.get('origin')
		});

		const corsHeaders = {
			'Access-Control-Allow-Origin': '*',
			'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
			'Access-Control-Allow-Headers': 'Content-Type, Authorization',
		};

		if (request.method === 'OPTIONS') {
			logMCPFeature('CORS_PREFLIGHT', { requestId, method: request.method });
			return new Response(null, { headers: corsHeaders });
		}

		if (url.pathname === '/' || url.pathname === '') {
			if (request.method === 'GET') {
				const duration = Date.now() - startTime;
				logMCPFeature('HEALTH_CHECK', { requestId, duration: `${duration}ms` });
				return new Response('✅ StreamNestAI MCP Server is running! Use POST / for MCP requests', {
					headers: { ...corsHeaders, 'Content-Type': 'text/plain' },
				});
			}

			if (request.method === 'POST') {
				try {
					const mcpRequest: MCPRequest = await request.json();
					
					logMCPFeature('MCP_REQUEST_RECEIVED', {
						requestId,
						method: mcpRequest.method,
						id: mcpRequest.id,
						hasParams: !!mcpRequest.params
					});

					const response = await handleMCPRequest(mcpRequest, BACKEND_URL, env, requestId);

					const duration = Date.now() - startTime;
					logMCPFeature('MCP_REQUEST_SUCCESS', {
						requestId,
						method: mcpRequest.method,
						duration: `${duration}ms`,
						hasResult: !!response.result,
						hasError: !!response.error
					});

					return new Response(JSON.stringify(response), {
						headers: {
							...corsHeaders,
							'Content-Type': 'application/json',
						},
					});
				} catch (error: any) {
					const duration = Date.now() - startTime;
					logMCPFeature('MCP_PARSE_ERROR', {
						requestId,
						duration: `${duration}ms`,
						error: error.message,
						stack: error.stack
					}, 'error');
					
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

		const duration = Date.now() - startTime;
		logMCPFeature('METHOD_NOT_ALLOWED', {
			requestId,
			method: request.method,
			path: url.pathname,
			duration: `${duration}ms`
		}, 'warn');

		return new Response('Method not allowed', { status: 405, headers: corsHeaders });
	},
};

async function handleMCPRequest(request: MCPRequest, BACKEND_URL: string, env: Env, requestId: string): Promise<MCPResponse> {
	const startTime = Date.now();
	const { method, params, id } = request;

	logMCPFeature('MCP_METHOD_START', {
		requestId,
		method,
		id,
		hasParams: !!params
	});

	try {
		if (method === 'initialize') {
			const duration = Date.now() - startTime;
			logMCPFeature('MCP_INITIALIZE', {
				requestId,
				duration: `${duration}ms`
			});
			
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
			const duration = Date.now() - startTime;
			logMCPFeature('MCP_TOOLS_LIST', {
				requestId,
				duration: `${duration}ms`,
				toolsCount: 5
			});
			
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

			logMCPFeature('TOOL_CALL_START', {
				requestId,
				toolName: name,
				hasArgs: !!args,
				argsKeys: args ? Object.keys(args) : []
			});

			const toolResult = await executeToolCall(name, args, BACKEND_URL, env.MOVIE_CACHE, requestId);

			const duration = Date.now() - startTime;
			logMCPFeature('TOOL_CALL_SUCCESS', {
				requestId,
				toolName: name,
				duration: `${duration}ms`,
				hasResult: !!toolResult,
				resultType: Array.isArray(toolResult) ? 'array' : typeof toolResult
			});

			return {
				jsonrpc: '2.0',
				id,
				result: {
					content: [{ type: 'text', text: JSON.stringify(toolResult, null, 2) }],
				},
			};
		}

		const duration = Date.now() - startTime;
		logMCPFeature('MCP_METHOD_NOT_FOUND', {
			requestId,
			method,
			duration: `${duration}ms`
		}, 'warn');

		return {
			jsonrpc: '2.0',
			id,
			error: { code: -32601, message: `Method not found: ${method}` },
		};
	} catch (error: any) {
		const duration = Date.now() - startTime;
		logMCPFeature('MCP_INTERNAL_ERROR', {
			requestId,
			method,
			duration: `${duration}ms`,
			error: error.message,
			stack: error.stack
		}, 'error');
		
		return {
			jsonrpc: '2.0',
			id,
			error: { code: -32603, message: 'Internal error', data: error.message },
		};
	}
}

async function executeToolCall(toolName: string, args: any, BACKEND_URL: string, CACHE: KVNamespace, requestId: string): Promise<any> {
	const startTime = Date.now();

	logMCPFeature('TOOL_EXECUTION_START', {
		requestId,
		toolName,
		hasArgs: !!args,
		argsKeys: args ? Object.keys(args) : []
	});

	switch (toolName) {
		case 'search_all_movies': {
			const cacheKey = 'all_movies';

			// Try cache first
			const cacheStartTime = Date.now();
			const cached = await CACHE.get(cacheKey, 'json');
			const cacheDuration = Date.now() - cacheStartTime;
			
			if (cached) {
				logMCPFeature('CACHE_HIT', {
					requestId,
					toolName,
					cacheKey,
					cacheDuration: `${cacheDuration}ms`,
					resultCount: Array.isArray(cached) ? cached.length : 'unknown'
				});
				return cached;
			}

			// Cache miss - fetch from backend
			logMCPFeature('CACHE_MISS', {
				requestId,
				toolName,
				cacheKey,
				cacheDuration: `${cacheDuration}ms`
			});

			const fetchStartTime = Date.now();
			const url = `${BACKEND_URL}/movies`;
			const res = await fetch(url);
			const fetchDuration = Date.now() - fetchStartTime;
			
			const text = await res.text();
			const data = JSON.parse(text);

			if (!Array.isArray(data)) {
				logMCPFeature('DATA_VALIDATION_ERROR', {
					requestId,
					toolName,
					expectedType: 'array',
					receivedType: typeof data,
					dataPreview: JSON.stringify(data).substring(0, 200)
				}, 'error');
				throw new Error(`Not an array`);
			}

			// Store in cache for 10 hours
			const storeStartTime = Date.now();
			await CACHE.put(cacheKey, JSON.stringify(data), {
				expirationTtl: 36000,
			});
			const storeDuration = Date.now() - storeStartTime;

			const totalDuration = Date.now() - startTime;
			logMCPFeature('CACHE_STORE_SUCCESS', {
				requestId,
				toolName,
				cacheKey,
				storeDuration: `${storeDuration}ms`,
				fetchDuration: `${fetchDuration}ms`,
				totalDuration: `${totalDuration}ms`,
				resultCount: data.length,
				ttl: 36000
			});
			
			return data;
		}

		case 'get_movie_by_id': {
			const { imdb_id } = args;
			const cacheKey = `movie_${imdb_id}`;

			logMCPFeature('MOVIE_LOOKUP_START', {
				requestId,
				toolName,
				imdb_id,
				cacheKey
			});

			// Try cache first
			const cacheStartTime = Date.now();
			const cached = await CACHE.get(cacheKey, 'json');
			const cacheDuration = Date.now() - cacheStartTime;
			
			if (cached) {
				logMCPFeature('CACHE_HIT', {
					requestId,
					toolName,
					cacheKey,
					imdb_id,
					cacheDuration: `${cacheDuration}ms`,
					hasData: !!cached
				});
				return cached;
			}

			// Cache miss
			logMCPFeature('CACHE_MISS', {
				requestId,
				toolName,
				cacheKey,
				imdb_id,
				cacheDuration: `${cacheDuration}ms`
			});

			const fetchStartTime = Date.now();
			const url = `${BACKEND_URL}/movies/${imdb_id}`;
			const res = await fetch(url);
			const fetchDuration = Date.now() - fetchStartTime;
			const data = await res.json();

			if (data.error || !res.ok) {
				logMCPFeature('MOVIE_NOT_FOUND', {
					requestId,
					toolName,
					imdb_id,
					status: res.status,
					statusText: res.statusText,
					fetchDuration: `${fetchDuration}ms`
				}, 'warn');
				throw new Error('Movie not found');
			}

			// Store in cache for 24 hours
			const storeStartTime = Date.now();
			await CACHE.put(cacheKey, JSON.stringify(data), {
				expirationTtl: 86400,
			});
			const storeDuration = Date.now() - storeStartTime;

			const totalDuration = Date.now() - startTime;
			logMCPFeature('CACHE_STORE_SUCCESS', {
				requestId,
				toolName,
				cacheKey,
				imdb_id,
				storeDuration: `${storeDuration}ms`,
				fetchDuration: `${fetchDuration}ms`,
				totalDuration: `${totalDuration}ms`,
				ttl: 86400
			});
			
			return data;
		}

		case 'get_recommended_movies': {
			logMCPFeature('RECOMMENDED_MOVIES_START', {
				requestId,
				toolName
			});

			const fetchStartTime = Date.now();
			const url = `${BACKEND_URL}/movies`;
			const res = await fetch(url);
			const fetchDuration = Date.now() - fetchStartTime;
			const text = await res.text();

			let data;
			try {
				data = JSON.parse(text);
			} catch (e) {
				logMCPFeature('JSON_PARSE_ERROR', {
					requestId,
					toolName,
					error: e.message,
					textPreview: text.substring(0, 100)
				}, 'error');
				throw new Error(`Invalid JSON: ${text.substring(0, 100)}`);
			}

			if (data.error) {
				logMCPFeature('BACKEND_ERROR', {
					requestId,
					toolName,
					error: data.error
				}, 'error');
				throw new Error(data.error);
			}

			if (!Array.isArray(data)) {
				logMCPFeature('DATA_VALIDATION_ERROR', {
					requestId,
					toolName,
					expectedType: 'array',
					receivedType: typeof data,
					dataPreview: JSON.stringify(data).substring(0, 200)
				}, 'error');
				throw new Error(`Backend returned non-array for movies: ${JSON.stringify(data).substring(0, 100)}`);
			}

			const processStartTime = Date.now();
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
			const processDuration = Date.now() - processStartTime;

			const totalDuration = Date.now() - startTime;
			logMCPFeature('RECOMMENDED_MOVIES_SUCCESS', {
				requestId,
				toolName,
				totalMovies: data.length,
				topRatedCount: topRated.length,
				fetchDuration: `${fetchDuration}ms`,
				processDuration: `${processDuration}ms`,
				totalDuration: `${totalDuration}ms`
			});

			if (topRated.length === 0) {
				return data.slice(0, 10);
			}

			return topRated;
		}

		case 'get_genres': {
			const cacheKey = 'genres';

			logMCPFeature('GENRES_START', {
				requestId,
				toolName,
				cacheKey
			});

			// Try cache first
			const cacheStartTime = Date.now();
			const cached = await CACHE.get(cacheKey, 'json');
			const cacheDuration = Date.now() - cacheStartTime;
			
			if (cached) {
				logMCPFeature('CACHE_HIT', {
					requestId,
					toolName,
					cacheKey,
					cacheDuration: `${cacheDuration}ms`,
					genresCount: Array.isArray(cached) ? cached.length : 'unknown'
				});
				return cached;
			}

			// Cache miss
			logMCPFeature('CACHE_MISS', {
				requestId,
				toolName,
				cacheKey,
				cacheDuration: `${cacheDuration}ms`
			});

			const fetchStartTime = Date.now();
			const url = `${BACKEND_URL}/genres`;
			const res = await fetch(url);
			const fetchDuration = Date.now() - fetchStartTime;
			const data = await res.json();

			if (!Array.isArray(data)) {
				logMCPFeature('DATA_VALIDATION_ERROR', {
					requestId,
					toolName,
					expectedType: 'array',
					receivedType: typeof data,
					dataPreview: JSON.stringify(data).substring(0, 200)
				}, 'error');
				throw new Error(`Not an array`);
			}

			// Store in cache for 24 hours
			const storeStartTime = Date.now();
			await CACHE.put(cacheKey, JSON.stringify(data), {
				expirationTtl: 86400,
			});
			const storeDuration = Date.now() - storeStartTime;

			const totalDuration = Date.now() - startTime;
			logMCPFeature('CACHE_STORE_SUCCESS', {
				requestId,
				toolName,
				cacheKey,
				storeDuration: `${storeDuration}ms`,
				fetchDuration: `${fetchDuration}ms`,
				totalDuration: `${totalDuration}ms`,
				genresCount: data.length,
				ttl: 86400
			});
			
			return data;
		}

		case 'search_by_keyword': {
			const { keyword } = args;
			const cacheKey = `search_${keyword.toLowerCase()}`;

			logMCPFeature('KEYWORD_SEARCH_START', {
				requestId,
				toolName,
				keyword,
				cacheKey
			});

			// Try cache first
			const cacheStartTime = Date.now();
			const cached = await CACHE.get(cacheKey, 'json');
			const cacheDuration = Date.now() - cacheStartTime;
			
			if (cached) {
				logMCPFeature('CACHE_HIT', {
					requestId,
					toolName,
					cacheKey,
					keyword,
					cacheDuration: `${cacheDuration}ms`,
					resultCount: Array.isArray(cached) ? cached.length : 'unknown'
				});
				return cached;
			}

			// Cache miss
			logMCPFeature('CACHE_MISS', {
				requestId,
				toolName,
				cacheKey,
				keyword,
				cacheDuration: `${cacheDuration}ms`
			});

			const fetchStartTime = Date.now();
			const url = `${BACKEND_URL}/movies`;
			const res = await fetch(url);
			const fetchDuration = Date.now() - fetchStartTime;

			if (!res.ok) {
				logMCPFeature('BACKEND_ERROR', {
					requestId,
					toolName,
					status: res.status,
					statusText: res.statusText,
					fetchDuration: `${fetchDuration}ms`
				}, 'error');
				throw new Error(`Backend returned ${res.status}: ${res.statusText}`);
			}

			const text = await res.text();

			let data;
			try {
				data = JSON.parse(text);
			} catch (e) {
				logMCPFeature('JSON_PARSE_ERROR', {
					requestId,
					toolName,
					error: e.message
				}, 'error');
				throw new Error(`Invalid JSON from backend`);
			}

			if (!Array.isArray(data)) {
				logMCPFeature('DATA_VALIDATION_ERROR', {
					requestId,
					toolName,
					expectedType: 'array',
					receivedType: typeof data
				}, 'error');
				throw new Error(`Backend returned non-array`);
			}

			const processStartTime = Date.now();
			const keywordLower = keyword.toLowerCase();

			// 🎯 SEARCHES ALL FIELDS automatically!
			const filtered = data.filter((m: any) => {
				// Convert entire movie to string and search
				const movieStr = JSON.stringify(m).toLowerCase();
				return movieStr.includes(keywordLower);
			});
			const processDuration = Date.now() - processStartTime;

			// Store search results for 100 minutes
			const storeStartTime = Date.now();
			await CACHE.put(cacheKey, JSON.stringify(filtered), {
				expirationTtl: 6000,
			});
			const storeDuration = Date.now() - storeStartTime;

			const totalDuration = Date.now() - startTime;
			logMCPFeature('KEYWORD_SEARCH_SUCCESS', {
				requestId,
				toolName,
				keyword,
				totalMovies: data.length,
				filteredCount: filtered.length,
				fetchDuration: `${fetchDuration}ms`,
				processDuration: `${processDuration}ms`,
				storeDuration: `${storeDuration}ms`,
				totalDuration: `${totalDuration}ms`,
				ttl: 6000
			});
			
			return filtered;
		}

		default:
			logMCPFeature('UNKNOWN_TOOL', {
				requestId,
				toolName,
				duration: `${Date.now() - startTime}ms`
			}, 'error');
			throw new Error(`Unknown tool: ${toolName}`);
	}
}
