import './App.css'
import logo from './images/logo.png';

function App() {
  return (
    <div className="container">
      <div className="bar">
        <h1 className="logotext">Brainwave</h1>
        <img className="logo" src={logo} alt="Logo" />
      </div>
      <p>TODO</p>
    </div>
  );
}

export default App;
