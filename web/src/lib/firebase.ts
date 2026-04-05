import { initializeApp } from 'firebase/app'
import {
  getAuth,
  GoogleAuthProvider,
  GithubAuthProvider,
  signInWithPopup,
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  signInWithPhoneNumber,
  RecaptchaVerifier,
  type ConfirmationResult,
} from 'firebase/auth'
import type { User as FirebaseUser } from 'firebase/auth'

const firebaseConfig = {
  apiKey: import.meta.env.VITE_FIREBASE_API_KEY,
  authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
  projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID,
}

const app = initializeApp(firebaseConfig)
export const firebaseAuth = getAuth(app)

const googleProvider = new GoogleAuthProvider()
const githubProvider = new GithubAuthProvider()

async function signInWithProvider(provider: GoogleAuthProvider | GithubAuthProvider): Promise<string> {
  const result = await signInWithPopup(firebaseAuth, provider)
  return result.user.getIdToken()
}

export async function signInWithGoogle(): Promise<string> {
  return signInWithProvider(googleProvider)
}

export async function signInWithGithub(): Promise<string> {
  return signInWithProvider(githubProvider)
}

export async function signInWithEmail(email: string, password: string): Promise<string> {
  const result = await signInWithEmailAndPassword(firebaseAuth, email, password)
  return result.user.getIdToken()
}

export async function signUpWithEmail(email: string, password: string): Promise<string> {
  const result = await createUserWithEmailAndPassword(firebaseAuth, email, password)
  return result.user.getIdToken()
}

// Phone auth — step 1: send OTP
// containerId must be the ID of a DOM element to anchor the invisible reCAPTCHA
export async function sendPhoneOTP(
  phoneNumber: string,
  containerId: string
): Promise<ConfirmationResult> {
  const verifier = new RecaptchaVerifier(firebaseAuth, containerId, { size: 'invisible' })
  return signInWithPhoneNumber(firebaseAuth, phoneNumber, verifier)
}

// Phone auth — step 2: confirm OTP, return Firebase ID token
export async function confirmPhoneOTP(
  confirmation: ConfirmationResult,
  otp: string
): Promise<string> {
  const result = await confirmation.confirm(otp)
  return result.user.getIdToken()
}

export type { FirebaseUser, ConfirmationResult }
