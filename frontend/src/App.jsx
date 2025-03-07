import './App.css'
import { BrowserRouter as Router, Route, Routes, Link } from 'react-router-dom';
import React, { useState } from 'react';
import {useNavigate } from 'react-router-dom';
import logo from './images/logo.png';

function HomePage() {
  return (
    <div className="container">
      <div className="bar">
        <h1 className="logotext">Brainwave</h1>
        <img className="logo" src={logo} alt="Logo" />
        <Link to="/login/">
          <button className="loginButton">Login</button>
        </Link>
      </div>
      <div className="subcontainer">
        <form className="searchBar"> 
         <input type="search" id="query" name="q" placeholder="Search..."></input>
         <button className="searchButton">Search</button>
        </form>
        <button className="recentButton">Recent</button>
        <button className="likesButton">Likes</button>
        <button className="oldestButton">Oldest</button>
        <div className="topicsContainer">
          
        </div>
      </div>
    </div>
  );
}

function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const [isModalOpen, setIsModalOpen] = useState(false); 
  
  const fakeUser = {
    username: 'testuser',
    password: 'password123',
  };

  const navigate = useNavigate();

  const handleLogin = (e) => {
    e.preventDefault(); 
 
    if (username === fakeUser.username && password === fakeUser.password) {
      navigate('/user/');
    } else {
      setErrorMessage('Invalid username or password. Please try again.');
      setIsModalOpen(true); 
    }
  };

  const closeModal = () => {
    setIsModalOpen(false); 
    setErrorMessage('');
  };

  return (
    <div className="container">
      <div className="bar">
        <h1 className="logotext">Brainwave</h1>
        <img className="logo" src={logo} alt="Logo" />
        <Link to="/">
          <button className="backButton">Back</button>
        </Link>
      </div>
      <div className="subcontainer">
        <div className="loginContainer">
          <input 
            className="usernameField" 
            placeholder="Username..." 
            value={username} 
            onChange={(e) => setUsername(e.target.value)} 
          />
          <input 
            className="passwordField" 
            type="password" 
            placeholder="Password..." 
            value={password} 
            onChange={(e) => setPassword(e.target.value)} 
          />
          <button 
            className="loginPageLoginButton"
            onClick={handleLogin} 
          >
            Login
          </button>
          <Link to="/login/askforemail/">
            <button className="forgotPasswordButton">Forgot Password?</button>
          </Link>
        </div>
      </div>

      {isModalOpen && (
        <div className="modal">
          <div className="modalContent">
            <p>{errorMessage}</p>
            <button className="closeModalButton" onClick={closeModal}>Close</button>
          </div>
        </div>
      )}
    </div>
  );
}

function AskForEmailPage() {
  return (
    <div className="container">
      <div className="bar">
        <h1 className="logotext">Brainwave</h1>
        <img className="logo" src={logo} alt="Logo" />
      </div>
      <div className="subcontainer">
        <div className="loginContainer">
          <p className="codeInstructions"> Type your email, a code will be sent to it so you can change your password</p>
          <input className="passwordField" placeholder="Type Your Email..."></input>
          <Link to="/login/askforemail/forgotpassword/">
            <button className="enterEmailButton">Change Password</button>
          </Link>
        </div>
      </div>
    </div>
  );
}

function ForgotPasswordPage() {
  return (
    <div className="container">
      <div className="bar">
        <h1 className="logotext">Brainwave</h1>
        <img className="logo" src={logo} alt="Logo" />
      </div>
      <div className="subcontainer">
        <div className="loginContainer">
          <p className="codeInstructions"> A code have been sent to your email, please type it and your new password of choice! You will be redirected to the mainpage after you are done.</p>
          <input className="passwordField" placeholder="Type New Password..."></input>
          <input className="codeField" placeholder="Type Your Code..."></input>
          <Link to="/">
            <button className="changePasswordButton">Proceed with password change</button>
          </Link>
        </div>
      </div>
    </div>
  );
}

function UserSettingPage() {
  return (
    <div className="container">
      <div className="bar">
        <h1 className="logotext">Brainwave</h1>
        <img className="logo" src={logo} alt="Logo" />
      </div>
      <div className="subcontainer">
      </div>
    </div>
  );
}

function UserFrontPage() {
  return (
    <div className="container">
      <div className="bar">
        <h1 className="logotext">Brainwave</h1>
        <img className="logo" src={logo} alt="Logo" />
        <Link to="/">
          <button className="logoutButton">Logout</button>
        </Link>
        <Link to="/user/account/">
          <button className="accountButton">Account</button>
        </Link>
      </div>
      <div className="subcontainer">
        <form className="searchBar"> 
         <input type="search" id="query" name="q" placeholder="Search..."></input>
         <button className="searchButton">Search</button>
        </form>
        <button className="recentButton">Recent</button>
        <button className="likesButton">Likes</button>
        <button className="oldestButton">Oldest</button>
        <div className="topicsContainer">
          
        </div>
      </div>
    </div>
  );
}

function TopicPage() {
  return (
    <div className="container">
      <div className="bar">
        <h1 className="logotext">Brainwave</h1>
        <img className="logo" src={logo} alt="Logo" />
      </div>
      <div className="subcontainer">
      </div>
    </div>
  );
}

function TopicPageUser() {
  return (
    <div className="container">
      <div className="bar">
        <h1 className="logotext">Brainwave</h1>
        <img className="logo" src={logo} alt="Logo" />
      </div>
      <div className="subcontainer">
      </div>
    </div>
  );
}

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/login/" element={<LoginPage />} />
        <Route path="/login/askforemail/" element={<AskForEmailPage />} />
        <Route path="/user/" element={<UserFrontPage />} />
        <Route path="/user/account/" element={<UserSettingPage />} />
        <Route path="/login/askforemail/forgotpassword" element={<ForgotPasswordPage />} />
      </Routes>
    </Router>
  );
}

export default App;
