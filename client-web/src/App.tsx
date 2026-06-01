import { Route, Routes } from "react-router";
import Header from "./components/Header";

function App() {
  return (
    <div className="flex flex-col h-screen w-full max-w-6xl">
      <Routes>
        <Route element={<Header />}>
          <Route path="/" element={<p>APP</p>} />
          <Route path="/login" element={<p>LOGIN</p>} />
          <Route path="/signup" element={<p>SIGNUP</p>} />
        </Route>
      </Routes>
    </div>
  );
}

export default App;
