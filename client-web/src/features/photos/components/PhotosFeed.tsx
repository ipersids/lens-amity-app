import type { Photo } from "../../../services/users";

const PhotosFeed = ({ photos }: { photos: Photo[] }) => {
  if (!photos.length) {
    return <p>No photos yet :)</p>;
  }

  return (
    <ul className="profile-photos" aria-label="Profile photos">
      {photos.map((photo) => (
        <li className="profile-photo" key={photo.photoID}>
          <a href={photo.request.url} aria-label={`${photo.title}, ${photo.date}`}>
            <img src={photo.request.url} alt={photo.title} />
            <span className="profile-photo-name">{photo.title}</span>
            <time dateTime={photo.date}>{photo.date}</time>
          </a>
        </li>
      ))}
    </ul>
  );
};

export default PhotosFeed;
