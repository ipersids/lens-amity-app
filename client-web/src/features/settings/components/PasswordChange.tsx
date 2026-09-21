import { XMarkIcon } from "@heroicons/react/24/outline";
import { useMutation } from "@tanstack/react-query";
import { Dialog } from "radix-ui";
import { useState } from "react";
import { getApiError } from "../../../services/api";
import usersService from "../../../services/users";
import PasswordField from "../../auth/components/PasswordField";
import { validatePassword } from "../../auth/validation";

const PasswordChange = ({ username }: { username: string }) => {
  const [open, setOpen] = useState<boolean>(false);
  const [revokeAll, setRevokeAll] = useState<boolean>(true);
  const [password, setPassword] = useState<string>("");
  const [newPassword, setNewPassword] = useState<string>("");

  const [error, setError] = useState<Partial<Record<"oldPassword" | "newPassword", string>>>({});

  const updatePassword = useMutation({
    mutationFn: usersService.updatePassword,
    onError: (error) => {
      const apiErr = getApiError(error);
      if (apiErr.error.code === "invalid_new_password") {
        setError({ newPassword: apiErr.error.message });
        return;
      }
      setError({ oldPassword: apiErr.error.message });
    },
    onSuccess: () => handleOpenChange(false),
  });

  const handleOpenChange = (isOpen: boolean) => {
    setOpen(isOpen);

    if (!isOpen) {
      setPassword("");
      setNewPassword("");
      setError({});
      setRevokeAll(true);
    }
  };

  const handleSaveChanges = (e: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
    e.stopPropagation();
    e.preventDefault();

    updatePassword.mutate({
      newPassword: newPassword,
      oldPassword: password,
      revokeAll: revokeAll,
    });
  };

  return (
    <Dialog.Root open={open} onOpenChange={handleOpenChange}>
      <div className="settings-account-action">
        <div>
          <h3>Change password</h3>
          <p>Update the password you use to sign in.</p>
        </div>
        <Dialog.Trigger asChild>
          <button className="dialog-trigger" type="button">
            Change password
          </button>
        </Dialog.Trigger>
        <Dialog.Portal>
          <Dialog.Overlay className="dialog-overlay" />
          <Dialog.Content className="dialog-content">
            <Dialog.Title className="dialog-title">Change password</Dialog.Title>
            <fieldset className="dialog-fieldset">
              <PasswordField
                id="current-password"
                autoComplete="current-password"
                name="current-password"
                onChange={(event) => {
                  setPassword(event.target.value);

                  setError((state) => ({
                    ...state,
                    oldPassword: undefined,
                  }));
                }}
                value={password}
                disabled={updatePassword.isPending}
                label="Old password"
                error={error.oldPassword}
              />

              <PasswordField
                id="new-password"
                autoComplete="new-password"
                name="new-password"
                onChange={(event) => {
                  const value = event.target.value;
                  setNewPassword(value);

                  const passwordValidation = validatePassword(value, [username]);
                  if (passwordValidation.feedback !== "") {
                    setError((state) => ({
                      ...state,
                      newPassword: passwordValidation.feedback,
                    }));
                  } else {
                    setError((state) => ({
                      ...state,
                      newPassword: undefined,
                    }));
                  }
                }}
                value={newPassword}
                disabled={updatePassword.isPending}
                label="New password"
                error={error.newPassword}
              />
            </fieldset>
            <div className="dialog-actions">
              <Dialog.Close asChild>
                <button className="dialog-cancel-button" type="button">
                  Cancel
                </button>
              </Dialog.Close>
              <Dialog.Close asChild>
                <button
                  className="dialog-save-button"
                  type="button"
                  onClick={(e) => handleSaveChanges(e)}
                  disabled={
                    !password ||
                    !newPassword ||
                    updatePassword.isPending ||
                    !!error.newPassword ||
                    !!error.oldPassword
                  }
                >
                  Save changes
                </button>
              </Dialog.Close>
            </div>
            <Dialog.Close asChild>
              <button className="dialog-close-button" type="button" aria-label="Close">
                <XMarkIcon />
              </button>
            </Dialog.Close>
          </Dialog.Content>
        </Dialog.Portal>
      </div>
    </Dialog.Root>
  );
};

export default PasswordChange;
