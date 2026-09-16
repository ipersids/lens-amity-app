import { Route, Routes } from "react-router";
import Layout from "./features/Layout";
import AddPhotoPage from "./pages/AddPhotoPage";
import AuthPage from "./pages/AuthPage";
import MyProfilePageRedirect from "./pages/MyProfilePageRedirect";
import ProfilePage from "./pages/ProfilePage";
import ProtectedPage from "./pages/ProtectedPage";
import SettingsPage from "./pages/SettingsPage";

function App() {
  return (
    <div className="app">
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<p>APP</p>} />
          <Route path="/login" element={<AuthPage mode="login" />} />
          <Route path="/signup" element={<AuthPage mode="signup" />} />

          <Route element={<ProtectedPage />}>
            <Route path="/profile" element={<MyProfilePageRedirect />} />
            <Route path="/settings" element={<SettingsPage />} />
            <Route path="/upload" element={<AddPhotoPage />} />
          </Route>

          <Route path="/users/:username" element={<ProfilePage />} />
        </Route>
      </Routes>
    </div>
  );
}

export default App;
