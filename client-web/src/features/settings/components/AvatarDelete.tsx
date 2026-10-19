import { XMarkIcon } from "@heroicons/react/24/outline";
import { Dialog } from "radix-ui";
import { useState } from "react";
import useDeleteAvatar from "../../../hooks/useDeleteAvatar";

const AvatarDelete = () => {
  const [open, setOpen] = useState(false);
  const deleteAvatar = useDeleteAvatar();

  const handleDeletClick = (e: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
    e.preventDefault();
    e.stopPropagation();
    deleteAvatar.mutate();

    setOpen(false);
  };

  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Trigger asChild>
        <button className="settings-avatar-button settings-avatar-delete-button" type="button">
          Delete
        </button>
      </Dialog.Trigger>

      <Dialog.Portal>
        <Dialog.Overlay className="dialog-overlay" />
        <Dialog.Content className="dialog-content">
          <Dialog.Title className="dialog-title">Delete profile photo?</Dialog.Title>
          <Dialog.Description className="dialog-description">
            Your current profile photo cannot be restored after deletion. Please be certain.
          </Dialog.Description>
          <div className="dialog-actions">
            <Dialog.Close asChild>
              <button
                className="dialog-cancel-button"
                type="button"
                disabled={deleteAvatar.isPending}
              >
                Cancel
              </button>
            </Dialog.Close>
            <button
              className="dialog-save-button dialog-delete-button"
              type="button"
              onClick={(event) => handleDeletClick(event)}
              disabled={deleteAvatar.isPending}
            >
              {deleteAvatar.isPending && <span className="button-spinner" aria-hidden="true" />}
              {deleteAvatar.isPending ? "Deleting..." : "Yes, delete profile photo"}
            </button>
          </div>
          <Dialog.Close asChild>
            <button
              className="dialog-close-button"
              type="button"
              aria-label="Close"
              disabled={deleteAvatar.isPending}
            >
              <XMarkIcon />
            </button>
          </Dialog.Close>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
};

export default AvatarDelete;
