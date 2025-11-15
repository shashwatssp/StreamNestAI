import { useState } from 'react';
import { Container, Row, Col, Form, Button, Card, Spinner, Alert } from 'react-bootstrap';
import Movies from '../movies/Movies';
import useAuth from '../../hooks/useAuth';
import './Explore.css';

const Explore = () => {
    const { auth } = useAuth();
    const [query, setQuery] = useState('');
    const [loading, setLoading] = useState(false);
    const [messages, setMessages] = useState([]);
    const [movies, setMovies] = useState([]);
    const [error, setError] = useState(null);

    const API_URL = import.meta.env.VITE_MCP_API_URL || 'https://streamnest-mcp.shashwatshagunpandey.workers.dev';

    // MCP client - manual implementation
const callMCPTool = async (toolName, args = {}) => {
    const response = await fetch(API_URL, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            jsonrpc: '2.0',
            id: Date.now(),
            method: 'tools/call',
            params: {
                name: toolName,
                arguments: args,
            },
        }),
    });

    const data = await response.json();
    
    // Better error logging
    if (data.error) {
        console.error('MCP Error:', data.error);
        console.error('Tool:', toolName);
        console.error('Args:', args);
        throw new Error(data.error.message + (data.error.data ? ': ' + data.error.data : ''));
    }
    
    return JSON.parse(data.result.content[0].text);
};


    const handleExplore = async (e) => {
        e.preventDefault();
        if (!query.trim()) return;

        setLoading(true);
        setError(null);
        setMovies([]);

        const userMessage = { role: 'user', content: query };
        setMessages(prev => [...prev, userMessage]);

        try {
            let result;

            if (query.toLowerCase().includes('recommend')) {
                result = await callMCPTool('get_recommended_movies', {
                    user_token: auth.token,
                });
                setMessages(prev => [...prev, {
                    role: 'assistant',
                    content: 'Here are your personalized recommendations!',
                }]);
            } else if (query.toLowerCase().includes('genre')) {
                result = await callMCPTool('get_genres', {});
                setMessages(prev => [...prev, {
                    role: 'assistant',
                    content: `Available genres: ${result.map(g => g.genre_name).join(', ')}`,
                }]);
                return; // Don't set movies for genres
            } else if (/tt\d+/.test(query)) {
                const imdbId = query.match(/tt\d+/)[0];
                result = await callMCPTool('get_movie_by_id', { imdb_id: imdbId });
                setMessages(prev => [...prev, {
                    role: 'assistant',
                    content: `Found "${result.title}"!`,
                }]);
                result = [result];
            } else {
                result = await callMCPTool('search_by_keyword', { keyword: query });
                setMessages(prev => [...prev, {
                    role: 'assistant',
                    content: `Found ${result.length} movies matching "${query}"`,
                }]);
            }

            setMovies(Array.isArray(result) ? result : [result]);
        } catch (err) {
            console.error('Error:', err);
            setError('Failed to process your request. Please try again.');
            setMessages(prev => [...prev, {
                role: 'assistant',
                content: 'Sorry, I encountered an error.',
            }]);
        } finally {
            setLoading(false);
        }
    };

    const quickPrompts = [
        { text: 'Show me my recommendations', icon: '🎯' },
        { text: 'action', icon: '💥' },
        { text: 'What are the available genres?', icon: '🎭' },
        { text: 'sci-fi', icon: '🌀' },
    ];

    const handleQuickPrompt = (promptText) => {
        setQuery(promptText);
    };

    return (
        <Container className="explore-container py-4">
            <Row className="mb-4">
                <Col>
                    <h2 className="text-center mb-2">🎬 Explore Movies with MCP</h2>
                    <p className="text-center text-muted">
                        Ask me anything about movies using Model Context Protocol!
                    </p>
                </Col>
            </Row>

            <Row className="mb-3">
                <Col>
                    <Alert variant="success">
                        ✅ MCP Server Connected - Ready to search!
                    </Alert>
                </Col>
            </Row>

            <Row className="mb-4">
                <Col md={10} className="mx-auto">
                    <Form onSubmit={handleExplore}>
                        <Form.Group className="mb-3">
                            <Form.Control
                                type="text"
                                placeholder="Try: 'action', 'Show me recommendations', or 'sci-fi'"
                                value={query}
                                onChange={(e) => setQuery(e.target.value)}
                                size="lg"
                            />
                        </Form.Group>
                        <Button 
                            variant="primary" 
                            type="submit" 
                            disabled={loading || !query.trim()}
                            className="w-100"
                            size="lg"
                        >
                            {loading ? (
                                <>
                                    <Spinner as="span" animation="border" size="sm" className="me-2" />
                                    Processing...
                                </>
                            ) : (
                                'Explore 🚀'
                            )}
                        </Button>
                    </Form>
                </Col>
            </Row>

            <Row className="mb-4">
                <Col md={10} className="mx-auto">
                    <p className="text-muted mb-2">Quick searches:</p>
                    <div className="d-flex flex-wrap gap-2">
                        {quickPrompts.map((prompt, idx) => (
                            <Button
                                key={idx}
                                variant="outline-secondary"
                                size="sm"
                                onClick={() => handleQuickPrompt(prompt.text)}
                                disabled={loading}
                            >
                                {prompt.icon} {prompt.text}
                            </Button>
                        ))}
                    </div>
                </Col>
            </Row>

            {messages.length > 0 && (
                <Row className="mb-4">
                    <Col md={10} className="mx-auto">
                        <Card className="chat-history">
                            <Card.Body>
                                {messages.map((msg, idx) => (
                                    <div key={idx} className={`message ${msg.role}`}>
                                        <strong>{msg.role === 'user' ? '👤 You' : '🤖 MCP'}:</strong>
                                        <span className="ms-2">{msg.content}</span>
                                    </div>
                                ))}
                            </Card.Body>
                        </Card>
                    </Col>
                </Row>
            )}

            {error && (
                <Row className="mb-3">
                    <Col md={10} className="mx-auto">
                        <Alert variant="danger" onClose={() => setError(null)} dismissible>
                            {error}
                        </Alert>
                    </Col>
                </Row>
            )}

            {loading ? (
                <div className="text-center">
                    <Spinner animation="border" variant="primary" />
                    <p className="mt-3">MCP is processing...</p>
                </div>
            ) : (
                movies.length > 0 && (
                    <Row>
                        <Col>
                            <Movies movies={movies} message="" />
                        </Col>
                    </Row>
                )
            )}
        </Container>
    );
};

export default Explore;
