package core

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"
	"log/slog"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nbutton23/zxcvbn-go"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// Docs:
// - NIST, SP 800-63B, authentication assurance: https://pages.nist.gov/800-63-4/sp800-63b.html
// - Unicode, Technical Standard #39, https://www.unicode.org/reports/tr39/#Restriction_Level_Detection

type AuthService struct {
	conf  *Auth
	store *db.Store
}

func NewAuthService(s *db.Store, confAuth *Auth) *AuthService {
	return &AuthService{
		conf:  confAuth,
		store: s,
	}
}

func (s *AuthService) ValidateAccessToken(tokenStr string) (*jwt.RegisteredClaims, error) {
	return validateToken(tokenStr, s.conf.JWTsecret)
}

var (
	usernameRegex         = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrCompromisedToken   = errors.New("token was compromised")
)

const dummyHash = "$argon2id$v=19$m=65536,t=3,p=2$72aaaaK2bbDJWl0/X2o4EQ$Nu9PSnVbhaHuKb5iLb6JDAdQ5z+0spTUEAO7tqBVvHA"

func (s *AuthService) Signup(ctx context.Context, uername, displayName, password string) (*db.CreateUserRow, error) {
	p := norm.NFC.String(password)
	ukey := normKey(uername)
	udisplay := normDisplay(displayName)

	if err := validateUsernameKey(ukey); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCredentials, err.Error())
	}

	if err := validatePassword(p, []string{ukey, udisplay}); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCredentials, err.Error())
	}

	if udisplay == "" {
		udisplay = normDisplay(uername)
	}

	if err := validateNameLength(udisplay); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCredentials, err.Error())
	}

	hash, err := generateFromPassword([]byte(p))
	if err != nil {
		return nil, err
	}

	user, err := s.store.Queries.CreateUser(ctx, db.CreateUserParams{
		UsernameKey:     ukey,
		PasswordHash:    hash,
		UsernameDisplay: udisplay,
	})

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, string, error) {
	p := norm.NFC.String(password)
	ukey := normKey(username)

	if err := validatePasswordLength(p); err != nil {
		return "", "", err
	}

	uPrivate, err := s.store.Queries.GetFullUserDataByKey(ctx, ukey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_ = compareHashAndPassword(dummyHash, []byte(p))
			return "", "", ErrInvalidCredentials
		}
		return "", "", err
	}

	if err := compareHashAndPassword(uPrivate.PasswordHash, []byte(p)); err != nil {
		return "", "", err
	}

	nowUTC := time.Now().UTC()

	token, err := s.signAccessToken(ctx, uPrivate.ID, nowUTC)
	if err != nil {
		return "", "", err
	}
	refreshTokenData, err := s.signRefreshToken(ctx, uPrivate.ID, nowUTC)
	if err != nil {
		return "", "", err
	}

	err = s.store.Queries.CreateNewRefreshToken(ctx, db.CreateNewRefreshTokenParams{
		ID:        refreshTokenData.id,
		UserID:    uPrivate.ID,
		ExpiresAt: refreshTokenData.claims.ExpiresAt.Time,
	})
	if err != nil {
		return "", "", err
	}

	return token, refreshTokenData.token, nil
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
	Replayed     bool
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	claims, err := validateToken(refreshToken, s.conf.RefreshSecret)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, err
	}

	now := time.Now().Truncate(time.Second)

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, err
	}

	tx, err := s.store.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rtoken, err := s.store.Queries.WithTx(tx).GetRefreshTokenForUpdate(ctx, db.GetRefreshTokenForUpdateParams{ID: tokenID, UserID: userID})
	if err != nil {
		return nil, err
	}

	if rtoken.Revoked {
		if rtoken.GracePeriodUntil.Valid && time.Now().Before(rtoken.GracePeriodUntil.Time) {
			return &RefreshResult{
				Replayed:     true,
				AccessToken:  "",
				RefreshToken: "",
			}, nil
		}
		// enable reuse detection
		if _, err := s.store.Queries.WithTx(tx).RevokeAllUserTokens(ctx, db.RevokeAllUserTokensParams{
			UserID:        userID,
			RevokedReason: db.NullTokenRevokedReason{TokenRevokedReason: db.TokenRevokedReasonReplayed, Valid: true},
		}); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrCompromisedToken, err.Error())
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrCompromisedToken, err.Error())
		}
		return nil, ErrCompromisedToken
	}

	accessToken, err := s.signAccessToken(ctx, userID, now)
	if err != nil {
		return nil, err
	}

	refreshTokenData, err := s.signRefreshToken(ctx, userID, now)
	if err != nil {
		return nil, err
	}

	_, err = s.store.Queries.WithTx(tx).RotateRefreshToken(ctx, db.RotateRefreshTokenParams{
		ID:               tokenID,
		UserID:           userID,
		GracePeriodUntil: pgtype.Timestamptz{Time: now.Add(3 * time.Second), Valid: true},
		RevokedReason:    db.NullTokenRevokedReason{TokenRevokedReason: db.TokenRevokedReasonRefresh, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	err = s.store.Queries.WithTx(tx).CreateNewRefreshToken(ctx, db.CreateNewRefreshTokenParams{
		ID:        refreshTokenData.id,
		UserID:    userID,
		ExpiresAt: refreshTokenData.claims.ExpiresAt.Time,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &RefreshResult{
		Replayed:     false,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenData.token,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	claims, err := validateToken(refreshToken, s.conf.RefreshSecret)
	if err != nil {
		slog.Error("1", "", err)
		return nil
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		slog.Error("2", "", err)
		return err
	}

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		return err
	}

	_, err = s.store.Queries.RevokeRefreshToken(ctx, db.RevokeRefreshTokenParams{
		RevokedReason: db.NullTokenRevokedReason{TokenRevokedReason: db.TokenRevokedReasonLogout, Valid: true},
		ID:            tokenID,
		UserID:        userID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	_, err := s.store.Queries.RevokeAllUserTokens(ctx, db.RevokeAllUserTokensParams{
		UserID:        userID,
		RevokedReason: db.NullTokenRevokedReason{TokenRevokedReason: db.TokenRevokedReasonLogout, Valid: true},
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	return nil
}

// ------------------------------ PRIVATE HELPERS ------------------------------

func normDisplay(s string) string {
	return norm.NFC.String(strings.TrimSpace(s))
}

func normKey(s string) string {
	folder := cases.Fold()
	s = strings.TrimSpace(s)
	s = norm.NFKC.String(s)
	s = folder.String(s)
	return s
}

func validatePassword(p string, userInputs []string) error {
	if err := validatePasswordLength(p); err != nil {
		return err
	}

	for _, r := range p {
		if !unicode.IsPrint(r) {
			return errors.New("input contains invalid characters")
		}
	}

	score := zxcvbn.PasswordStrength(p, userInputs)

	if score.Score < 3 {
		return errors.New("password is too weak and easy to guess")
	}

	// @TODO: compare password against breach/common-password blocklist

	return nil
}

func validatePasswordLength(p string) error {
	l := utf8.RuneCountInString(p)

	// NIST, SP 800-63B for single-factor authentication:
	// - minimum of 15 characters in length
	if l < 15 {
		return errors.New("password must be at least 15 characters")
	}
	// - maximum at least 64
	if l > 128 {
		return errors.New("password too long")
	}

	return nil
}

func validateNameLength(name string) error {
	l := utf8.RuneCountInString(name)

	if l < 3 {
		return errors.New("display name is too short")
	}

	if l > 32 {
		return errors.New("display name is too long")
	}

	return nil
}

// Apply strict rules for username key
func validateUsernameKey(u string) error {
	if !usernameRegex.MatchString(u) {
		return errors.New("username must be 3-32 characters and only contain alphanumeric characters, underscores, or hyphens")
	}

	if reservedUsernames[u] {
		return errors.New("username is not available")
	}

	return nil
}

// ReservedUsernames contains paths and terms that users are banned from registering
// https://github.com/shouldbee/reserved-usernames
var reservedUsernames = map[string]bool{
	"about":            true,
	"access":           true,
	"account":          true,
	"accounts":         true,
	"activate":         true,
	"activities":       true,
	"activity":         true,
	"add":              true,
	"address":          true,
	"adm":              true,
	"admin":            true,
	"administration":   true,
	"administrator":    true,
	"ads":              true,
	"adult":            true,
	"advertising":      true,
	"affiliate":        true,
	"affiliates":       true,
	"ajax":             true,
	"all":              true,
	"alpha":            true,
	"analysis":         true,
	"analytics":        true,
	"android":          true,
	"anon":             true,
	"anonymous":        true,
	"api":              true,
	"app":              true,
	"apps":             true,
	"archive":          true,
	"archives":         true,
	"article":          true,
	"asct":             true,
	"asset":            true,
	"atom":             true,
	"auth":             true,
	"authentication":   true,
	"avatar":           true,
	"backup":           true,
	"balancer-manager": true,
	"banner":           true,
	"banners":          true,
	"beta":             true,
	"billing":          true,
	"bin":              true,
	"blog":             true,
	"blogs":            true,
	"board":            true,
	"book":             true,
	"bookmark":         true,
	"bot":              true,
	"bots":             true,
	"bug":              true,
	"business":         true,
	"cache":            true,
	"cadastro":         true,
	"calendar":         true,
	"call":             true,
	"campaign":         true,
	"cancel":           true,
	"captcha":          true,
	"career":           true,
	"careers":          true,
	"cart":             true,
	"categories":       true,
	"category":         true,
	"cgi":              true,
	"cgi-bin":          true,
	"changelog":        true,
	"chat":             true,
	"check":            true,
	"checking":         true,
	"checkout":         true,
	"client":           true,
	"cliente":          true,
	"clients":          true,
	"code":             true,
	"codereview":       true,
	"comercial":        true,
	"comment":          true,
	"comments":         true,
	"communities":      true,
	"community":        true,
	"company":          true,
	"compare":          true,
	"compras":          true,
	"config":           true,
	"configuration":    true,
	"connect":          true,
	"contact":          true,
	"contact-us":       true,
	"contact_us":       true,
	"contactus":        true,
	"contest":          true,
	"contribute":       true,
	"corp":             true,
	"create":           true,
	"css":              true,
	"dashboard":        true,
	"data":             true,
	"default":          true,
	"delete":           true,
	"demo":             true,
	"design":           true,
	"designer":         true,
	"destroy":          true,
	"dev":              true,
	"devel":            true,
	"developer":        true,
	"developers":       true,
	"diagram":          true,
	"diary":            true,
	"dict":             true,
	"dictionary":       true,
	"die":              true,
	"dir":              true,
	"direct_messages":  true,
	"directory":        true,
	"dist":             true,
	"doc":              true,
	"docs":             true,
	"documentation":    true,
	"domain":           true,
	"download":         true,
	"downloads":        true,
	"ecommerce":        true,
	"edit":             true,
	"editor":           true,
	"edu":              true,
	"education":        true,
	"email":            true,
	"employment":       true,
	"empty":            true,
	"end":              true,
	"enterprise":       true,
	"entries":          true,
	"entry":            true,
	"error":            true,
	"errors":           true,
	"eval":             true,
	"event":            true,
	"everyone":         true,
	"exit":             true,
	"explore":          true,
	"facebook":         true,
	"faq":              true,
	"favorite":         true,
	"favorites":        true,
	"feature":          true,
	"features":         true,
	"feed":             true,
	"feedback":         true,
	"feeds":            true,
	"file":             true,
	"files":            true,
	"first":            true,
	"flash":            true,
	"fleet":            true,
	"fleets":           true,
	"flog":             true,
	"follow":           true,
	"followers":        true,
	"following":        true,
	"forgot":           true,
	"form":             true,
	"forum":            true,
	"forums":           true,
	"founder":          true,
	"free":             true,
	"friend":           true,
	"friends":          true,
	"ftp":              true,
	"gadget":           true,
	"gadgets":          true,
	"game":             true,
	"games":            true,
	"get":              true,
	"ghost":            true,
	"gift":             true,
	"gifts":            true,
	"gist":             true,
	"github":           true,
	"graph":            true,
	"group":            true,
	"groups":           true,
	"guest":            true,
	"guests":           true,
	"help":             true,
	"home":             true,
	"homepage":         true,
	"host":             true,
	"hosting":          true,
	"hostmaster":       true,
	"hostname":         true,
	"howto":            true,
	"hpg":              true,
	"html":             true,
	"http":             true,
	"httpd":            true,
	"https":            true,
	"iamges":           true,
	"icon":             true,
	"icons":            true,
	"idea":             true,
	"ideas":            true,
	"image":            true,
	"images":           true,
	"imap":             true,
	"img":              true,
	"index":            true,
	"indice":           true,
	"info":             true,
	"information":      true,
	"inquiry":          true,
	"instagram":        true,
	"intranet":         true,
	"invitations":      true,
	"invite":           true,
	"ipad":             true,
	"iphone":           true,
	"irc":              true,
	"issue":            true,
	"issues":           true,
	"item":             true,
	"items":            true,
	"java":             true,
	"javascript":       true,
	"job":              true,
	"jobs":             true,
	"join":             true,
	"json":             true,
	"jump":             true,
	"knowledgebase":    true,
	"language":         true,
	"languages":        true,
	"last":             true,
	"ldap-status":      true,
	"legal":            true,
	"license":          true,
	"link":             true,
	"links":            true,
	"linux":            true,
	"list":             true,
	"lists":            true,
	"log":              true,
	"log-in":           true,
	"log-out":          true,
	"log_in":           true,
	"log_out":          true,
	"login":            true,
	"logout":           true,
	"logs":             true,
	"mac":              true,
	"mail":             true,
	"mail1":            true,
	"mail2":            true,
	"mail3":            true,
	"mail4":            true,
	"mail5":            true,
	"mailer":           true,
	"mailing":          true,
	"maintenance":      true,
	"manager":          true,
	"manual":           true,
	"map":              true,
	"maps":             true,
	"marketing":        true,
	"master":           true,
	"media":            true,
	"member":           true,
	"members":          true,
	"message":          true,
	"messages":         true,
	"messenger":        true,
	"microblog":        true,
	"microblogs":       true,
	"mine":             true,
	"mis":              true,
	"mob":              true,
	"mobile":           true,
	"movie":            true,
	"movies":           true,
	"mp3":              true,
	"msg":              true,
	"msn":              true,
	"music":            true,
	"musicas":          true,
	"mysql":            true,
	"name":             true,
	"named":            true,
	"nan":              true,
	"navi":             true,
	"navigation":       true,
	"net":              true,
	"network":          true,
	"new":              true,
	"news":             true,
	"newsletter":       true,
	"nick":             true,
	"nickname":         true,
	"notes":            true,
	"noticias":         true,
	"notification":     true,
	"notifications":    true,
	"notify":           true,
	"ns1":              true,
	"ns10":             true,
	"ns2":              true,
	"ns3":              true,
	"ns4":              true,
	"ns5":              true,
	"ns6":              true,
	"ns7":              true,
	"ns8":              true,
	"ns9":              true,
	"null":             true,
	"oauth":            true,
	"oauth_clients":    true,
	"offer":            true,
	"offers":           true,
	"official":         true,
	"old":              true,
	"online":           true,
	"openid":           true,
	"operator":         true,
	"order":            true,
	"orders":           true,
	"organization":     true,
	"organizations":    true,
	"overview":         true,
	"owner":            true,
	"owners":           true,
	"page":             true,
	"pager":            true,
	"pages":            true,
	"panel":            true,
	"password":         true,
	"payment":          true,
	"perl":             true,
	"phone":            true,
	"photo":            true,
	"photoalbum":       true,
	"photos":           true,
	"php":              true,
	"phpmyadmin":       true,
	"phppgadmin":       true,
	"phpredisadmin":    true,
	"pic":              true,
	"pics":             true,
	"ping":             true,
	"plan":             true,
	"plans":            true,
	"plugin":           true,
	"plugins":          true,
	"policy":           true,
	"pop":              true,
	"pop3":             true,
	"popular":          true,
	"portal":           true,
	"post":             true,
	"postfix":          true,
	"postmaster":       true,
	"posts":            true,
	"premium":          true,
	"press":            true,
	"price":            true,
	"pricing":          true,
	"privacy":          true,
	"privacy-policy":   true,
	"privacy_policy":   true,
	"privacypolicy":    true,
	"private":          true,
	"product":          true,
	"products":         true,
	"profile":          true,
	"project":          true,
	"projects":         true,
	"promo":            true,
	"pub":              true,
	"public":           true,
	"purpose":          true,
	"put":              true,
	"python":           true,
	"query":            true,
	"random":           true,
	"ranking":          true,
	"read":             true,
	"readme":           true,
	"recent":           true,
	"recruit":          true,
	"recruitment":      true,
	"register":         true,
	"registration":     true,
	"release":          true,
	"remove":           true,
	"replies":          true,
	"report":           true,
	"reports":          true,
	"repositories":     true,
	"repository":       true,
	"req":              true,
	"request":          true,
	"requests":         true,
	"reset":            true,
	"roc":              true,
	"root":             true,
	"rss":              true,
	"ruby":             true,
	"rule":             true,
	"sag":              true,
	"sale":             true,
	"sales":            true,
	"sample":           true,
	"samples":          true,
	"save":             true,
	"school":           true,
	"script":           true,
	"scripts":          true,
	"search":           true,
	"secure":           true,
	"security":         true,
	"self":             true,
	"send":             true,
	"server":           true,
	"server-info":      true,
	"server-status":    true,
	"service":          true,
	"services":         true,
	"session":          true,
	"sessions":         true,
	"setting":          true,
	"settings":         true,
	"setup":            true,
	"share":            true,
	"shop":             true,
	"show":             true,
	"sign-in":          true,
	"sign-up":          true,
	"sign_in":          true,
	"sign_up":          true,
	"signin":           true,
	"signout":          true,
	"signup":           true,
	"site":             true,
	"sitemap":          true,
	"sites":            true,
	"smartphone":       true,
	"smtp":             true,
	"soporte":          true,
	"source":           true,
	"spec":             true,
	"special":          true,
	"sql":              true,
	"src":              true,
	"ssh":              true,
	"ssl":              true,
	"ssladmin":         true,
	"ssladministrator": true,
	"sslwebmaster":     true,
	"staff":            true,
	"stage":            true,
	"staging":          true,
	"start":            true,
	"stat":             true,
	"state":            true,
	"static":           true,
	"stats":            true,
	"status":           true,
	"store":            true,
	"stores":           true,
	"stories":          true,
	"style":            true,
	"styleguide":       true,
	"stylesheet":       true,
	"stylesheets":      true,
	"subdomain":        true,
	"subscribe":        true,
	"subscriptions":    true,
	"suporte":          true,
	"support":          true,
	"svn":              true,
	"swf":              true,
	"sys":              true,
	"sysadmin":         true,
	"sysadministrator": true,
	"system":           true,
	"tablet":           true,
	"tablets":          true,
	"tag":              true,
	"talk":             true,
	"task":             true,
	"tasks":            true,
	"team":             true,
	"teams":            true,
	"tech":             true,
	"telnet":           true,
	"term":             true,
	"terms":            true,
	"terms-of-service": true,
	"terms_of_service": true,
	"termsofservice":   true,
	"test":             true,
	"test1":            true,
	"test2":            true,
	"test3":            true,
	"teste":            true,
	"testing":          true,
	"tests":            true,
	"theme":            true,
	"themes":           true,
	"thread":           true,
	"threads":          true,
	"tmp":              true,
	"todo":             true,
	"tool":             true,
	"tools":            true,
	"top":              true,
	"topic":            true,
	"topics":           true,
	"tos":              true,
	"tour":             true,
	"translations":     true,
	"trends":           true,
	"tutorial":         true,
	"tux":              true,
	"tv":               true,
	"twitter":          true,
	"undef":            true,
	"unfollow":         true,
	"unsubscribe":      true,
	"update":           true,
	"upload":           true,
	"uploads":          true,
	"url":              true,
	"usage":            true,
	"user":             true,
	"username":         true,
	"users":            true,
	"usuario":          true,
	"vendas":           true,
	"ver":              true,
	"version":          true,
	"video":            true,
	"videos":           true,
	"visitor":          true,
	"watch":            true,
	"weather":          true,
	"web":              true,
	"webhook":          true,
	"webhooks":         true,
	"webmail":          true,
	"webmaster":        true,
	"website":          true,
	"websites":         true,
	"welcome":          true,
	"widget":           true,
	"widgets":          true,
	"wiki":             true,
	"win":              true,
	"windows":          true,
	"word":             true,
	"work":             true,
	"works":            true,
	"workshop":         true,
	"wws":              true,
	"www":              true,
	"www1":             true,
	"www2":             true,
	"www3":             true,
	"www4":             true,
	"www5":             true,
	"www6":             true,
	"www7":             true,
	"wwws":             true,
	"wwww":             true,
	"xfn":              true,
	"xml":              true,
	"xmpp":             true,
	"xpg":              true,
	"xxx":              true,
	"yaml":             true,
	"year":             true,
	"yml":              true,
	"you":              true,
	"lensamity":        true,
	"lens-amity":       true,
	"lens_amity":       true,
}
