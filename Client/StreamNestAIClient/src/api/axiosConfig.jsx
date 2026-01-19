import axios from 'axios'
const apiURL = import.meta.env.VITE_API_BASE_URL

// Create axios instance with enhanced logging
const axiosClient = axios.create({
    baseURL: apiURL,
    headers: {'Content-Type':'application/json'},
    withCredentials: true // Important for HTTP-only cookies
})

// Request interceptor for logging
axiosClient.interceptors.request.use(
    (config) => {
        const timestamp = new Date().toISOString()
        console.log(`🔵 [FRONTEND API] ${timestamp} - ${config.method?.toUpperCase()} ${config.url}`)
        console.log(`🔵 [FRONTEND API] Request headers:`, config.headers)
        if (config.data) {
            console.log(`🔵 [FRONTEND API] Request data:`, config.data)
        }
        if (config.params) {
            console.log(`🔵 [FRONTEND API] Request params:`, config.params)
        }
        console.log(`🔵 [FRONTEND API] Full URL: ${config.baseURL}${config.url}`)
        return config
    },
    (error) => {
        const timestamp = new Date().toISOString()
        console.error(`🔴 [FRONTEND API] ${timestamp} - Request Error:`, error)
        return Promise.reject(error)
    }
)

// Response interceptor for logging
axiosClient.interceptors.response.use(
    (response) => {
        const timestamp = new Date().toISOString()
        const duration = response.config.metadata?.endTime ?
            `${response.config.metadata.endTime - response.config.metadata.startTime}ms` :
            'unknown'
        
        console.log(`🟢 [FRONTEND API] ${timestamp} - Response ${response.status} ${response.statusText} (${duration})`)
        console.log(`🟢 [FRONTEND API] Response headers:`, response.headers)
        
        // Log response data size for large responses
        if (response.data) {
            const dataSize = JSON.stringify(response.data).length
            console.log(`🟢 [FRONTEND API] Response data size: ${dataSize} bytes`)
            
            // For array responses, log count
            if (Array.isArray(response.data)) {
                console.log(`🟢 [FRONTEND API] Response array length: ${response.data.length} items`)
            }
            
            // Log first few items for debugging (but not too much)
            if (Array.isArray(response.data) && response.data.length > 0) {
                console.log(`🟢 [FRONTEND API] First item sample:`, response.data[0])
            } else if (!Array.isArray(response.data)) {
                console.log(`🟢 [FRONTEND API] Response data sample:`, response.data)
            }
        }
        
        return response
    },
    (error) => {
        const timestamp = new Date().toISOString()
        console.error(`🔴 [FRONTEND API] ${timestamp} - Response Error:`, error)
        
        if (error.response) {
            // The request was made and the server responded with a status code
            // that falls out of the range of 2xx
            console.error(`🔴 [FRONTEND API] Error Status: ${error.response.status} ${error.response.statusText}`)
            console.error(`🔴 [FRONTEND API] Error Data:`, error.response.data)
            console.error(`🔴 [FRONTEND API] Error Headers:`, error.response.headers)
        } else if (error.request) {
            // The request was made but no response was received
            console.error(`🔴 [FRONTEND API] No Response Received:`, error.request)
        } else {
            // Something happened in setting up the request that triggered an Error
            console.error(`🔴 [FRONTEND API] Request Setup Error:`, error.message)
        }
        
        return Promise.reject(error)
    }
)

// Add timing metadata to requests
axiosClient.interceptors.request.use((config) => {
    config.metadata = { startTime: new Date().getTime() }
    return config
})

axiosClient.interceptors.response.use((response) => {
    response.config.metadata.endTime = new Date().getTime()
    return response
})

export default axiosClient