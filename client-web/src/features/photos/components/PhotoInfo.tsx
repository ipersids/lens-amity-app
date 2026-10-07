import {
  ArrowLeftIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { Link } from "react-router";
import usePhotoID from "../../../hooks/usePhotoID";
import { formatPhotoDate } from "../../../utils";
import { ProfileNotAvailable } from "../../profile";
import Avatar from "../../profile/components/Avatar";

type UserParams = {
  username: string;
  displayName: string;
  canEdit: boolean;
  avatar?: {
    url: string;
  };
};

type PhotoInfoParams = {
  profile: UserParams;
  photoID: string;
};

const PhotoInfo = ({ profile, photoID }: PhotoInfoParams) => {
  const { isPending, isError, data } = usePhotoID(profile.username, photoID);

  if (isError) {
    return <ProfileNotAvailable />;
  }

  if (isPending || !data) {
    return null;
  }

  return (
    <article className="photo-info" data-photo-id={data.photo.photoID}>
      <header className="photo-info-header">
        <Link className="photo-info-back" to={`/users/${profile.username}`}>
          <ArrowLeftIcon aria-hidden="true" />
          Back to profile
        </Link>

        {profile.canEdit && (
          <button aria-label="Delete photo" className="photo-info-delete" type="button">
            <TrashIcon aria-hidden="true" />
          </button>
        )}
      </header>

      <div className="photo-info-card">
        <section aria-label="Photo preview" className="photo-info-preview">
          <img src={data.photo.request.url} alt={data.photo.title} />

          {data.previousPhotoID && (
            <Link
              aria-label="Previous photo"
              className="photo-info-arrow photo-info-arrow--previous"
              to={`/users/${encodeURIComponent(profile.username)}/photos/${encodeURIComponent(data.previousPhotoID)}`}
            >
              <ChevronLeftIcon aria-hidden="true" />
            </Link>
          )}
          {data.nextPhotoID && (
            <Link
              aria-label="Next photo"
              className="photo-info-arrow photo-info-arrow--next"
              to={`/users/${encodeURIComponent(profile.username)}/photos/${encodeURIComponent(data.nextPhotoID)}`}
            >
              <ChevronRightIcon aria-hidden="true" />
            </Link>
          )}
        </section>

        <section className="photo-info-details">
          <div className="photo-info-author">
            <Avatar
              avatarURL={profile.avatar?.url}
              displayName={profile.displayName}
              size="settings"
            />
            <div className="photo-info-author-copy">
              <span>Photo by</span>
              <Link to={`/users/${profile.username}`}>{profile.displayName}</Link>
            </div>
          </div>

          <div className="photo-info-copy">
            <time dateTime={data.photo.date}>{formatPhotoDate(data.photo.date)}</time>
            <h1>{data.photo.title}</h1>
            <p>{data.photo.description}</p>
          </div>
        </section>
      </div>
    </article>
  );
};

export default PhotoInfo;
