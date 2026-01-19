import {useState} from 'react'
import Button from 'react-bootstrap/Button'
import Container from 'react-bootstrap/Container'
import Nav from 'react-bootstrap/Nav'
import Navbar from 'react-bootstrap/Navbar'
import NavDropdown from 'react-bootstrap/NavDropdown'
import {useNavigate, NavLink, Link} from 'react-router-dom'
import useAuth from '../../hooks/useAuth';
import logo from '../../assets/StreamNestAILogo.png';

const Header = ({handleLogout}) => {
    const navigate = useNavigate();
    const {auth} = useAuth();


    return (
        <Navbar bg="dark" variant='dark' expand="lg" stick="top" className="shadow-sm">
            <Container>
                <Navbar.Brand>
                     <img
                        alt=""
                        src={logo}
                        width="30"
                        height="30"
                        className="d-inline-block align-top me-2"
                    />
                    StreamNestAI
                </Navbar.Brand>

            <Navbar.Toggle aria-controls="main-navbar-nav" />
                <Navbar.Collapse>
                    <Nav className ="me-auto">
                        <Nav.Link as = {NavLink} to="/">
                            Home
                        </Nav.Link>
                        <Nav.Link as = {NavLink} to="/recommended">
                            Recommended
                        </Nav.Link>
                        <Nav.Link as = {NavLink} to="/explore">
                            ✨ Explore
                        </Nav.Link>
                        
                        {/* New Features Dropdown */}
                        <NavDropdown title="🚀 Features" id="features-nav-dropdown">
                            <NavDropdown.Item as = {NavLink} to="/social">
                                👥 Social Features
                            </NavDropdown.Item>
                            <NavDropdown.Item as = {NavLink} to="/recommendations">
                                🎯 Advanced Recommendations
                            </NavDropdown.Item>
                            <NavDropdown.Item as = {NavLink} to="/watchlist">
                                📝 Watchlist Manager
                            </NavDropdown.Item>
                            <NavDropdown.Item as = {NavLink} to="/search">
                                🔍 Natural Search
                            </NavDropdown.Item>
                            <NavDropdown.Divider />
                            <NavDropdown.Item as = {NavLink} to="/profile-enhanced">
                                👤 Enhanced Profile
                            </NavDropdown.Item>
                        </NavDropdown>

                        {/* Admin & Analytics Dropdown */}
                        <NavDropdown title="📊 Dashboard" id="dashboard-nav-dropdown">
                            <NavDropdown.Item as = {NavLink} to="/analytics">
                                📈 Analytics Dashboard
                            </NavDropdown.Item>
                            <NavDropdown.Item as = {NavLink} to="/monitoring">
                                🖥️ System Monitoring
                            </NavDropdown.Item>
                            <NavDropdown.Divider />
                            <NavDropdown.Item as = {NavLink} to="/admin">
                                ⚙️ Admin Panel
                            </NavDropdown.Item>
                        </NavDropdown>
                    </Nav>
    
                    <Nav className ="ms-auto align-items-center" style={{padding: '10px'}}>
                        {auth ? (
                        <>  <div style={{padding:'0px 10px'}} >

                        
                            <span className="me-3 text-light">
                                Hello, <strong>{auth.first_name}</strong>
                            </span>
                            <Button variant="outline-light" size="sm" onClick={handleLogout}>
                                Logout
                            </Button>
                            </div>
                        </>
                        ):(
                            <div style={{padding:'0px 15px'}}>  
                                <Button
                                    variant="outline-info"
                                    size="sm"
                                    className="me-2"
                                    onClick={() => navigate("/login")}
                                    style={{padding: '0px 15px'}}
                                >
                                    Login
                                </Button>
                                
                                <Button
                                    variant="info"
                                    size="sm"
                                    onClick={() => navigate("/register")}
                                    style={{padding: '0px 15px'}}
                                >
                                    Register
                                </Button>                        
                            </div>
                        )}
                    </Nav>       
                </Navbar.Collapse>
            </Container>
        </Navbar>
    )
}
export default Header;