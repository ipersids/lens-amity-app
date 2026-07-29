import { Route, Routes } from "react-router";
import Layout from "./features/Layout";
import AddPhotoPage from "./pages/AddPhotoPage";
import AuthPage from "./pages/AuthPage";
import ProtectedPage from "./pages/ProtectedPage";

function App() {
  return (
    <div className="app">
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<p>APP</p>} />
          <Route path="/login" element={<AuthPage mode="login" />} />
          <Route path="/signup" element={<AuthPage mode="signup" />} />
          <Route element={<ProtectedPage />}>
            <Route path="/upload" element={<AddPhotoPage />} />
          </Route>
        </Route>
      </Routes>
    </div>
  );
}

export default App;
