import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ProfileQueryKey } from "../constants";
import type { UploadAvatarIntentItem } from "../services/avatar";
import avatarService from "../services/avatar";
import { useUser } from "../stores/auth";

type UpdateAvatarVariables = {
  avatarInfo: UploadAvatarIntentItem;
  file: File;
};

const useUpdateAvatar = () => {
  const currentUser = useUser();
  const queryClient = useQueryClient();
  const profileKey = ProfileQueryKey(currentUser?.username ?? "");

  return useMutation({
    mutationFn: ({ avatarInfo, file }: UpdateAvatarVariables) =>
      avatarService.upload(avatarInfo, file),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: profileKey, exact: true }),
  });
};

export default useUpdateAvatar;
