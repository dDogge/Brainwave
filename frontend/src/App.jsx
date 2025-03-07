import './App.css';
import { BrowserRouter as Router, Route, Routes, Link, useParams, useNavigate } from 'react-router-dom';
import React, { useState } from 'react';
import logo from './images/logo.png';

const fakeTopics = [
  { id: 1, title: "Fake Topic One", likes: 10, timestamp: '2023-04-01T10:00:00Z' },
  { id: 2, title: "Fake Topic Two", likes: 5, timestamp: '2023-03-15T12:00:00Z' },
  { id: 3, title: "Fake Topic Three", likes: 20, timestamp: '2023-05-10T09:00:00Z' },
];

function TopicsList({ topics, userPage, sortMode }) {
  // Sortera topics baserat på valt sorteringsläge
  const sortedTopics = topics.slice().sort((a, b) => {
    if (sortMode === 'likes') {
      return b.likes - a.likes;
    } else if (sortMode === 'recent') {
      return new Date(b.timestamp) - new Date(a.timestamp);
    } else if (sortMode === 'oldest') {
      return new Date(a.timestamp) - new Date(b.timestamp);
    } else {
      return 0;
    }
  });

  return (
    <div className="topicsList">
      {sortedTopics.map(topic => (
        <Link key={topic.id} to={userPage ? `/user/topic/${topic.id}` : `/topic/${topic.id}`}>
          <div className="topicItem">
            <h3>{topic.title}</h3>
            <p>{topic.likes} likes</p>
            <p>{new Date(topic.timestamp).toLocaleString()}</p>
          </div>
        </Link>
      ))}
    </div>
  );
}

function HomePage() {
  const [sortMode, setSortMode] = useState('recent');

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
          <input type="search" id="query" name="q" placeholder="Search..." />
          <button className="searchButton">Search</button>
        </form>
        <button className="recentButton" onClick={() => setSortMode('recent')}>Recent</button>
        <button className="likesButton" onClick={() => setSortMode('likes')}>Likes</button>
        <button className="oldestButton" onClick={() => setSortMode('oldest')}>Oldest</button>
        <div className="topicsContainer">
          <TopicsList topics={fakeTopics} userPage={false} sortMode={sortMode} />
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
  const [sortMode, setSortMode] = useState('recent');

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
          <input type="search" id="query" name="q" placeholder="Search..." />
          <button className="searchButton">Search</button>
        </form>
        <button className="recentButton" onClick={() => setSortMode('recent')}>Recent</button>
        <button className="likesButton" onClick={() => setSortMode('likes')}>Likes</button>
        <button className="oldestButton" onClick={() => setSortMode('oldest')}>Oldest</button>
        <div className="topicsContainer">
          <TopicsList topics={fakeTopics} userPage={true} sortMode={sortMode} />
        </div>
      </div>
    </div>
  );
}

function TopicPage() {
  const { id } = useParams();
  return (
    <div className="container">
      <div className="bar">
        <h1 className="logotext">Brainwave</h1>
        <img className="logo" src={logo} alt="Logo" />
      </div>
      <div className="subcontainer">
        <h2>Topic {id}</h2>
        <p>This is the topic page for non-logged in users.</p>
      </div>
    </div>
  );
}

function TopicPageUser() {
  const { id } = useParams();
  return (
    <div className="container">
      <div className="bar">
        <h1 className="logotext">Brainwave</h1>
        <img className="logo" src={logo} alt="Logo" />
      </div>
      <div className="subcontainer">
        <h2>Topic {id}</h2>
        <p>This is the topic page for logged in users.</p>
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
        <Route path="/login/askforemail/forgotpassword" element={<ForgotPasswordPage />} />
        <Route path="/user/" element={<UserFrontPage />} />
        <Route path="/user/account/" element={<UserSettingPage />} />
        <Route path="/topic/:id" element={<TopicPage />} />
        <Route path="/user/topic/:id" element={<TopicPageUser />} />
      </Routes>
    </Router>
  );
}

export default App;
