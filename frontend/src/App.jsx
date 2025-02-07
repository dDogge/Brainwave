import './App.css'
import { BrowserRouter as Router, Route, Routes, Link } from 'react-router-dom';
import logo from './images/logo.png';

function HomePage() {
  return (
    <div className="container">
      <div className="bar">
        <h1 className="logotext">Brainwave</h1>
        <img className="logo" src={logo} alt="Logo" />
        <Link to="/login">
          <button className="loginButton">login</button>
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

function ForgotPasswordPage() {
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
        <Route path="/login" element={<LoginPage />} />
      </Routes>
    </Router>
  );
}

export default App;
