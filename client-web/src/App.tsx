import { Route, Routes } from "react-router";
import Layout from "./components/Layout";

function App() {
  return (
    <div className="app">
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<p>APP</p>} />
          <Route path="/login" element={<p>LOGIN</p>} />
          <Route path="/signup" element={<p>SIGNUP</p>} />
        </Route>
      </Routes>
    </div>
  );
}

export default App;
