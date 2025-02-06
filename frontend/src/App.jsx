import './App.css'
import logo from './images/logo.png';

function App() {
  return (
    <div className="container">
      <div className="bar">
        <h1 className="logotext">Brainwave</h1>
        <img className="logo" src={logo} alt="Logo" />
        <button className="loginButton">login</button>
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

export default App;
