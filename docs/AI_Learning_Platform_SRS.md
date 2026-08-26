# Software Requirements Specification (SRS)
## AI-Powered Personalized Learning Platform

*Derived from the approved Business/Product Blueprint. No database schema or infrastructure choices are made here — those are deferred to the architecture stage.*

---

## 1. Introduction

### 1.1 System Purpose
The system shall convert a learner-defined goal into a personalized, evidence-verified learning path, built from expert-vetted knowledge structures and curated external resources, and shall continuously adapt that path based on demonstrated competency rather than content consumption.

### 1.2 Scope
The system covers: goal capture, learner baselining, expert knowledge-structure management, resource curation, personalized path generation, learning delivery, evidence-based assessment, competency tracking, adaptive remediation, progress/explainability reporting, and the administrative tooling curators need to maintain the above.

### 1.3 System Boundaries
The system does **not** author primary learning content, does not operate as a general-purpose chatbot, does not freely generate curricula without an expert-defined base, does not issue certifications at this stage, and does not include social/community or marketplace functionality.

### 1.4 Intended Users
Learners (primary), Content Curators/Domain Experts (maintain knowledge structures and resource quality), and Platform Administrators (system-level operations, user and access management).

### 1.5 Major Capabilities
Goal definition and learner profiling; diagnostic assessment and skill-gap identification; expert roadmap management; resource discovery and curation; personalized roadmap generation; learning delivery; assessment (quiz/scenario/project); competency evaluation; adaptive remediation; progress tracking and explainability; notifications; feedback capture; administrative controls.

### 1.6 Out-of-Scope Functionality (v1)
In-house content authoring/CMS beyond curation metadata; social or community features; expert marketplace; certification issuance; recruiter-facing profiles; database schema design; cloud/infrastructure selection; specific ML/algorithm selection.

---

## 2. Actors (User Roles)

### 2.1 Prospective Learner (Guest)
- **Responsibilities:** Explore the product, define a candidate goal, optionally complete a baseline preview.
- **Permissions:** Read-only access to public product information; may initiate but not persist a full profile without registering.
- **System interactions:** Goal input, baseline preview.
- **Data accessible:** None persisted beyond the session unless conversion to a registered Learner occurs.
- **Data modifiable:** None.

### 2.2 Learner
- **Responsibilities:** Define goals, complete baseline and ongoing assessments, engage with recommended resources, submit evidence of understanding.
- **Permissions:** Full read/write access to their own profile, goals, path, assessment history, and competency record. No access to other learners' data.
- **System interactions:** All learner-facing capabilities in Section 1.5 except administrative/curation functions.
- **Data accessible:** Own profile, own path, own competency record, resource metadata (not other learners' data).
- **Data modifiable:** Own profile, own goals/preferences, own assessment submissions.

### 2.3 Content Curator (Domain Expert / Admin)
- **Responsibilities:** Author and maintain expert knowledge structures (concepts, prerequisites, sequencing constraints); vet and curate resources; review flagged or stale resources; review ambiguous assessment cases when automated grading is insufficient.
- **Permissions:** Read/write access to knowledge structures and resource catalog for assigned domains; read access to aggregated (not individually identifying beyond necessity) learner competency data for quality feedback loops.
- **System interactions:** Roadmap management tools, resource curation tools, resource/knowledge-structure review queues.
- **Data accessible:** Knowledge structures, resource catalog, aggregated competency/usefulness metrics, individual assessment artifacts only when handling an escalated review.
- **Data modifiable:** Knowledge structures and resource catalog for their assigned scope, subject to platform review rules.

### 2.4 Platform Administrator
- **Responsibilities:** User and access management, platform configuration, oversight of curator activity, incident response, audit review.
- **Permissions:** Full administrative access to system configuration, user accounts (not learner content itself beyond what's needed for support), and audit logs.
- **System interactions:** Admin console, audit log review, curator/role management.
- **Data accessible:** Account metadata, system configuration, audit logs, aggregated system metrics.
- **Data modifiable:** Account status/roles, system configuration.

### 2.5 Automated Assessment/Adaptation Engine (system actor, not human)
Referenced throughout Functional Requirements as the internal process responsible for scoring evidence, updating competency records, and computing adaptive path changes. Included here for completeness of interaction modeling; not a user-facing role.

---

## 3. Functional Requirements

Priority key: **M** = Must (MVP), **S** = Should (MVP+), **L** = Later.

### 3.1 Authentication

**FR-AUTH-01 — Learner Registration & Login**
- *Description:* The system shall allow a Prospective Learner to register an account and authenticate as a Learner.
- *Preconditions:* None (new user) / valid credentials exist (returning user).
- *Inputs:* Identity credentials (e.g., email + password or supported SSO).
- *Processing:* Validate credentials, create or retrieve account, establish authenticated session.
- *Outputs:* Authenticated session; profile record created on first registration.
- *Success condition:* Learner is authenticated and routed to onboarding (new) or dashboard (returning).
- *Failure condition:* Invalid credentials, duplicate account on registration, or unverified account block login with a clear, non-revealing error.
- *Priority:* M

**FR-AUTH-02 — Session Management**
- *Description:* The system shall maintain a secure session for authenticated users and expire it after inactivity or explicit logout.
- *Preconditions:* Successful authentication.
- *Inputs:* Session token, activity events.
- *Processing:* Validate token per request; expire per policy.
- *Outputs:* Continued access or forced re-authentication.
- *Success condition:* Session remains valid only within policy window and is invalidated on logout.
- *Failure condition:* Expired/invalid token blocks access and prompts re-login.
- *Priority:* M

### 3.2 User Onboarding

**FR-ONB-01 — Guided Onboarding Flow**
- *Description:* The system shall guide a new Learner through goal capture, baseline assessment, and preference collection before generating an initial path.
- *Preconditions:* Learner is registered.
- *Inputs:* Goal statement, baseline assessment responses, learning preferences (time available, format preference, constraints).
- *Processing:* Sequence the above steps; persist partial progress if interrupted.
- *Outputs:* Learner profile populated; readiness flag for path generation.
- *Success condition:* All required onboarding data captured and profile marked ready for path generation.
- *Failure condition:* Incomplete required data blocks path generation with a clear resumable state.
- *Priority:* M

### 3.3 Goal Definition

**FR-GOAL-01 — Define/Update Learning Goal**
- *Description:* The system shall allow a Learner to define a specific learning/career goal and later revise it.
- *Preconditions:* Learner authenticated.
- *Inputs:* Free-text or structured goal input, mapped to an available expert knowledge structure.
- *Processing:* Validate the goal against known knowledge structures; if unmatched, flag as unsupported (see Edge Cases).
- *Outputs:* Goal record linked to a knowledge structure; triggers baseline/re-baseline if goal changes materially.
- *Success condition:* Goal is recognized and mapped to a valid knowledge structure.
- *Failure condition:* Goal has no matching knowledge structure; system informs learner rather than inventing an unvetted path.
- *Priority:* M

### 3.4 Learner Profiling

**FR-PROF-01 — Maintain Learner Profile**
- *Description:* The system shall maintain a profile capturing constraints (time, budget if relevant), preferences (resource format), and declared prior experience.
- *Preconditions:* Learner authenticated.
- *Inputs:* Learner-provided profile data; updates over time.
- *Processing:* Store and version profile changes; make available to path generation and resource selection logic.
- *Outputs:* Current profile state.
- *Success condition:* Profile reflects latest learner-provided data and is used in subsequent personalization.
- *Failure condition:* Missing required profile fields degrade personalization gracefully to sensible defaults, not to a blocked state.
- *Priority:* M

### 3.5 Skill Representation

**FR-SKILL-01 — Represent Concepts and Competency States**
- *Description:* The system shall represent each concept within a knowledge structure and track a per-learner, per-concept competency state (e.g., Not Started / In Progress / Weak Evidence / Competent).
- *Preconditions:* Knowledge structure exists for the relevant domain.
- *Inputs:* Assessment evidence, curator-defined concept definitions.
- *Processing:* Map evidence outcomes to competency state transitions per defined rules (see Business Rules, Section 5).
- *Outputs:* Per-learner competency record per concept.
- *Success condition:* Competency state accurately reflects the most recent qualifying evidence.
- *Failure condition:* No evidence recorded leaves state at Not Started/In Progress; state must never advance without qualifying evidence.
- *Priority:* M

### 3.6 Diagnostic Assessment

**FR-DIAG-01 — Baseline Diagnostic**
- *Description:* The system shall administer a diagnostic assessment during onboarding (and on demand later) to estimate a learner's existing competency across relevant concepts.
- *Preconditions:* Goal defined and mapped to a knowledge structure.
- *Inputs:* Learner responses to diagnostic items tied to concepts in the knowledge structure.
- *Processing:* Score responses; populate initial competency estimates.
- *Outputs:* Initial per-concept competency estimates.
- *Success condition:* Baseline competency estimates exist for all concepts relevant to the goal.
- *Failure condition:* Incomplete diagnostic yields partial/low-confidence estimates flagged as such, not treated as full competency.
- *Priority:* M

### 3.7 Skill-Gap Identification

**FR-GAP-01 — Identify Gaps Against Goal**
- *Description:* The system shall compute the difference between required competency (per the goal's knowledge structure) and the learner's current competency record.
- *Preconditions:* Baseline diagnostic complete; knowledge structure available.
- *Inputs:* Current competency record, target knowledge structure requirements.
- *Processing:* Compute per-concept gap; prioritize gaps respecting prerequisite order.
- *Outputs:* Ordered list of concept gaps driving path generation.
- *Success condition:* Gap list is consistent with prerequisite constraints from the knowledge structure.
- *Failure condition:* Conflicting or missing prerequisite data is flagged for curator review rather than silently resolved.
- *Priority:* M

### 3.8 Expert Roadmap Management

**FR-EXPERT-01 — Author/Maintain Knowledge Structures**
- *Description:* The system shall allow Content Curators to create and maintain knowledge structures: concepts, prerequisite relationships, and sequencing constraints per domain.
- *Preconditions:* Curator authenticated with domain scope.
- *Inputs:* Concept definitions, prerequisite mappings, versioning metadata.
- *Processing:* Validate internal consistency (e.g., no circular prerequisites); version changes.
- *Outputs:* Published or draft knowledge structure version.
- *Success condition:* Structure passes consistency validation and is available for path generation.
- *Failure condition:* Inconsistent structure (e.g., circular dependency) is rejected with diagnostic detail.
- *Priority:* M

### 3.9 Resource Discovery

**FR-DISC-01 — Discover Candidate Resources**
- *Description:* The system shall support discovery of candidate external resources mapped to specific concepts, for curator review before publication.
- *Preconditions:* Concept exists in a knowledge structure.
- *Inputs:* Candidate resource metadata (source, format, url), concept mapping.
- *Processing:* Stage candidates for curator vetting; capture provenance metadata at discovery time.
- *Outputs:* Candidate resource queue.
- *Success condition:* Discovered candidates are queued with provenance intact.
- *Failure condition:* Candidates missing provenance are not auto-published.
- *Priority:* S

### 3.10 Gold-Standard Resource Curation

**FR-CUR-01 — Vet and Publish Resources**
- *Description:* The system shall allow Curators to vet, approve, and publish resources for a concept, and to retire or flag resources over time.
- *Preconditions:* Candidate resource exists.
- *Inputs:* Curator review decision, quality/freshness metadata.
- *Processing:* Record decision, provenance, and review date; schedule freshness re-check.
- *Outputs:* Published resource available for personalization; retired resources removed from active rotation.
- *Success condition:* Only vetted, provenance-complete resources are servable to learners.
- *Failure condition:* Resource failing freshness re-check is flagged and withheld pending review (see Business Rules).
- *Priority:* M

### 3.11 Personalized Roadmap Generation

**FR-PATH-01 — Generate Personalized Path**
- *Description:* The system shall generate a learner-specific sequence of concepts and resources from the gap list, the knowledge structure's valid constraints, and the learner's profile/preferences.
- *Preconditions:* Gap list computed; published resources available for relevant concepts.
- *Inputs:* Gap list, knowledge structure constraints, learner profile.
- *Processing:* Sequence concepts respecting prerequisites; select best-fit resource per concept per preferences.
- *Outputs:* Personalized path instance.
- *Success condition:* Path respects all prerequisite constraints and uses only published resources.
- *Failure condition:* No valid resource for a required concept halts path generation for that concept and surfaces the gap to curators, rather than substituting an unvetted resource.
- *Priority:* M

### 3.12 Learning Experience Delivery

**FR-LEARN-01 — Deliver Concept Content**
- *Description:* The system shall present the next concept and its associated resource(s) to the learner within their active path.
- *Preconditions:* Personalized path exists.
- *Inputs:* Path state, learner navigation actions.
- *Processing:* Render resource reference/content and track engagement at a basic level (not as a competency signal).
- *Outputs:* Concept/resource presented to learner.
- *Success condition:* Learner can access the assigned resource and proceed to the associated assessment.
- *Failure condition:* Resource unreachable triggers fallback resource selection or curator flag (see Edge Cases).
- *Priority:* M

### 3.13 Quiz/Assessment

**FR-QUIZ-01 — Administer Concept Assessment**
- *Description:* The system shall administer an assessment (quiz, scenario, or other evidence-eliciting format) for a concept after the learner engages with its resource(s).
- *Preconditions:* Concept resource delivered.
- *Inputs:* Learner responses.
- *Processing:* Score responses against concept-specific criteria.
- *Outputs:* Assessment result feeding competency evaluation.
- *Success condition:* Result is recorded and available to the competency engine.
- *Failure condition:* Incomplete submission does not generate a competency-advancing result.
- *Priority:* M

### 3.14 Project Submission

**FR-PROJ-01 — Submit Practical/Project Evidence**
- *Description:* The system shall allow a Learner to submit practical task or project evidence for concepts requiring applied demonstration.
- *Preconditions:* Concept requires project-level evidence per knowledge structure.
- *Inputs:* Learner submission (artifact, description, or structured response).
- *Processing:* Route to automated evaluation and/or curator review as defined for that concept.
- *Outputs:* Evaluation result feeding competency evaluation.
- *Success condition:* Submission is evaluated and a result recorded.
- *Failure condition:* Evaluation inconclusive routes to human curator review rather than defaulting to pass or fail.
- *Priority:* S

### 3.15 Competency Evaluation

**FR-COMP-01 — Update Competency Record**
- *Description:* The system shall update a learner's per-concept competency state based solely on recorded assessment/project evidence.
- *Preconditions:* New evidence recorded (FR-QUIZ-01 or FR-PROJ-01).
- *Inputs:* Evidence result(s).
- *Processing:* Apply competency-state transition rules; version the change with a timestamp and evidence reference.
- *Outputs:* Updated competency record.
- *Success condition:* State change is fully traceable to specific evidence.
- *Failure condition:* Any state change lacking a traceable evidence reference is rejected.
- *Priority:* M

### 3.16 Remediation

**FR-REM-01 — Trigger Targeted Remediation**
- *Description:* The system shall trigger remediation for a specific weak concept when evidence indicates insufficient competency, rather than restarting a broader module.
- *Preconditions:* Competency evaluation marks a concept as weak/failed.
- *Inputs:* Weak-concept identification, prior attempt history.
- *Processing:* Select remediation resource(s) targeted at the specific weak concept; may vary from the originally assigned resource.
- *Outputs:* Remediation step inserted into the learner's active path.
- *Success condition:* Remediation targets only the identified weak concept(s).
- *Failure condition:* Repeated remediation failure on the same concept escalates for curator review rather than looping indefinitely.
- *Priority:* M

### 3.17 Adaptive Decisions

**FR-ADAPT-01 — Recompute Path on New Evidence**
- *Description:* The system shall re-evaluate and, where needed, adjust the remaining path whenever new competency evidence is recorded.
- *Preconditions:* Competency record updated.
- *Inputs:* Updated competency record, current path state.
- *Processing:* Recompute gap list and downstream sequencing; preserve prerequisite validity.
- *Outputs:* Updated path instance.
- *Success condition:* Path reflects current competency without violating knowledge-structure constraints.
- *Failure condition:* Recomputation that would violate a prerequisite constraint is rejected and logged for review.
- *Priority:* M

### 3.18 Progress Tracking

**FR-PROG-01 — Track and Display Progress**
- *Description:* The system shall track and display a learner's progress toward their goal in terms of competency achieved, not content consumed.
- *Preconditions:* Active path exists.
- *Inputs:* Competency record, path state.
- *Processing:* Aggregate competency-based progress metrics.
- *Outputs:* Progress view for the learner.
- *Success condition:* Displayed progress is derived only from competency evidence.
- *Failure condition:* No content-consumption-only metric (e.g., "% watched") is presented as progress.
- *Priority:* M

### 3.19 Explainability

**FR-EXPL-01 — Explain Path/Resource Decisions**
- *Description:* The system shall provide a learner-facing explanation of why a given concept or resource was selected next.
- *Preconditions:* Path/resource decision exists.
- *Inputs:* Decision context (gap, prerequisite state, preference match).
- *Processing:* Generate a human-readable rationale referencing the actual decision factors.
- *Outputs:* Explanation shown on request.
- *Success condition:* Explanation accurately reflects the actual decision inputs.
- *Failure condition:* No explanation is fabricated if the underlying rationale cannot be reconstructed; system discloses the limitation instead.
- *Priority:* S

### 3.20 Notifications

**FR-NOTIF-01 — Learner Notifications**
- *Description:* The system shall notify learners of relevant events (remediation triggered, path updated, assessment available, goal milestone reached).
- *Preconditions:* Notification-worthy event occurs.
- *Inputs:* Event type, learner notification preferences.
- *Processing:* Match event to enabled channel(s); dispatch.
- *Outputs:* Delivered notification.
- *Success condition:* Notification reflects an actual system event and respects learner preferences.
- *Failure condition:* Delivery failure is retried per policy and does not block underlying system state changes.
- *Priority:* S

### 3.21 Feedback

**FR-FDBK-01 — Capture Resource/Path Feedback**
- *Description:* The system shall allow learners to rate or flag resources and path decisions.
- *Preconditions:* Learner has engaged with the resource/decision.
- *Inputs:* Rating, free-text flag/comment.
- *Processing:* Store feedback; surface aggregated feedback to curators for resource quality review.
- *Outputs:* Feedback record; aggregated quality signal.
- *Success condition:* Feedback is attributable for curator follow-up without being treated as competency evidence.
- *Failure condition:* Feedback volume/severity crossing a threshold flags the resource for curator review.
- *Priority:* S

### 3.22 Administrative Controls

**FR-ADMIN-01 — Manage Users, Roles, and System Configuration**
- *Description:* The system shall allow Platform Administrators to manage user accounts, assign curator/admin roles, and configure system-level parameters.
- *Preconditions:* Administrator authenticated.
- *Inputs:* Role assignments, configuration changes.
- *Processing:* Apply changes; log the change with actor and timestamp.
- *Outputs:* Updated account/role/configuration state.
- *Success condition:* Change is applied and fully attributable in the audit log.
- *Failure condition:* Unauthorized or malformed change request is rejected.
- *Priority:* M

---

## 4. Non-Functional Requirements

| Category | MVP Target | Production Target |
|---|---|---|
| Performance | Core learner actions (view path, submit assessment) respond within a few seconds under light load | Consistent responsiveness under expected concurrent user load, with defined load-testing benchmarks |
| Latency | Non-blocking UX for path/competency recomputation (may run asynchronously) | Near-real-time recomputation for most learners; async fallback for heavy recomputation |
| Availability | Best-effort availability during business hours; planned maintenance windows acceptable | High-availability target with defined uptime commitment and monitored SLOs |
| Scalability | Support an initial cohort of learners in a small number of domains | Horizontally scalable to many domains and a large concurrent learner base |
| Security | Authenticated access, role-based permissions, encrypted credentials | Full security program: encryption in transit/at rest, regular audits, vulnerability management |
| Privacy | Learner data accessible only to the learner and authorized roles; no cross-learner data exposure | Formal privacy policy compliance, data minimization, configurable data retention/deletion |
| Reliability | Core competency/evidence data must not be lost on transient failures | Defined recovery point/time objectives; redundancy for critical data paths |
| Maintainability | Knowledge structures and resources editable without code changes | Versioned, auditable change management for structures and resources |
| Observability | Basic logging of key state transitions (competency updates, path changes) | Full monitoring/alerting on system health and business-critical flows |
| Accessibility | Baseline accessible UI patterns (readable content, keyboard navigable core flows) | Formal accessibility compliance target (e.g., WCAG-aligned) |
| Auditability | Curator and admin actions logged with actor/timestamp | Immutable audit trail with retention policy and review tooling |
| Data Integrity | Competency state changes always traceable to specific evidence | Enforced data validation and consistency checks across all evidence-to-state transitions |
| Fault Tolerance | Graceful degradation when a resource or external dependency is unavailable (see Edge Cases) | Defined fallback and recovery strategy per dependency, with monitoring |

---

## 5. Business Rules

- **BR-01:** The system shall not allow a learner to unlock a dependent concept until sufficient competency evidence exists for its prerequisite concept(s), as defined by the knowledge structure.
- **BR-02:** Curriculum structure — concepts, prerequisites, and sequencing constraints — shall originate only from expert-authored/vetted knowledge structures. No component of the system (including any AI-assisted feature) may freely invent or alter prerequisite relationships.
- **BR-03:** Every resource servable to a learner must have recorded provenance: source, author/publisher where known, and curation/review date.
- **BR-04:** Resource quality and freshness must be tracked; resources exceeding a defined staleness threshold must be withheld from active rotation until re-reviewed.
- **BR-05:** Personalization (sequencing, resource selection, pacing) may vary within the valid constraints of the expert-defined knowledge structure but must never violate its prerequisite/dependency rules.
- **BR-06:** Competency status for a concept must be derived only from recorded assessment/project evidence — never from content-consumption signals alone (e.g., time spent, percentage viewed).
- **BR-07:** Remediation triggered by a weak-area determination must target the specific weak concept(s) identified, not a broad re-do of unrelated material.
- **BR-08:** A learner's recorded competency state must not change without a traceable evidence reference; any downgrade requires new evidence and must be explainable on request.
- **BR-09:** The system must be able to produce, on request, an explanation of why a given concept or resource was selected next, grounded in the actual decision inputs.
- **BR-10:** If a knowledge structure is updated while a learner has an in-progress path built on a prior version, the learner's in-progress path shall remain consistent with the version it was generated from unless the learner explicitly opts into a refresh.
- **BR-11:** Business-model tier restrictions (e.g., free vs. paid) may limit path depth/breadth but must never misrepresent a learner's actual competency status.
- **BR-12:** All curator and administrator changes to knowledge structures, resources, roles, or configuration must be attributable to an actor and timestamp for audit purposes.

---

## 6. Use Cases

**UC-01 — New Learner Onboarding**
*Actor:* Prospective Learner → Learner. *Trigger:* Registration. *Precondition:* None. *Main flow:* Register → define goal → complete baseline diagnostic → capture preferences → system generates initial path. *Exception:* Goal unmatched to any knowledge structure → learner informed, goal flagged for curator consideration. *Postcondition:* Learner has an active personalized path.

**UC-02 — Existing Learner Returning**
*Actor:* Learner. *Trigger:* Login. *Precondition:* Registered account. *Main flow:* Authenticate → view current path/progress → resume next concept. *Exception:* Session expired → re-authenticate. *Postcondition:* Learner resumes at correct path state.

**UC-03 — Goal Creation/Update**
*Actor:* Learner. *Trigger:* New or revised goal input. *Precondition:* Authenticated. *Main flow:* Submit goal → system maps to knowledge structure → gap analysis re-run → path (re)generated. *Exception:* Unsupported goal → learner informed. *Postcondition:* Goal recorded and linked to an active or updated path.

**UC-04 — Diagnostic Assessment**
*Actor:* Learner. *Trigger:* Onboarding or on-demand re-baseline. *Precondition:* Goal defined. *Main flow:* Present diagnostic items → learner responds → system scores and sets initial competency estimates. *Exception:* Incomplete diagnostic → partial/low-confidence estimates flagged. *Postcondition:* Baseline competency established.

**UC-05 — Personalized Roadmap Generation**
*Actor:* System (triggered by Learner action). *Trigger:* Gap list available. *Precondition:* Published resources exist for required concepts. *Main flow:* Compute gaps → sequence per prerequisites → select best-fit resources → publish path. *Exception:* Missing resource for a required concept → concept surfaced as a gap, curator notified. *Postcondition:* Learner has a valid, personalized path.

**UC-06 — Learning a Concept**
*Actor:* Learner. *Trigger:* Learner opens next path item. *Precondition:* Active path exists. *Main flow:* Present resource → learner engages → learner proceeds to assessment. *Exception:* Resource unreachable → fallback resource or curator flag. *Postcondition:* Learner ready for assessment.

**UC-07 — Taking an Assessment**
*Actor:* Learner. *Trigger:* Concept resource completed. *Precondition:* Assessment defined for the concept. *Main flow:* Present assessment → learner submits → system scores → result recorded. *Exception:* Incomplete submission → no competency-advancing result generated. *Postcondition:* Evidence recorded for competency evaluation.

**UC-08 — Competency Update**
*Actor:* System. *Trigger:* New evidence recorded. *Precondition:* Assessment/project result exists. *Main flow:* Apply state-transition rules → update competency record → recompute gaps. *Exception:* Evidence lacks traceability → change rejected. *Postcondition:* Competency record reflects latest evidence.

**UC-09 — Failure/Remediation**
*Actor:* Learner, System. *Trigger:* Competency evaluation marks a concept weak/failed. *Precondition:* Assessment result recorded. *Main flow:* System identifies weak concept → selects targeted remediation resource → inserts into path. *Exception:* Repeated failure on same concept → escalate to curator review. *Postcondition:* Learner has a remediation step targeting the actual gap.

**UC-10 — Resource Recommendation**
*Actor:* System. *Trigger:* Path generation or adaptation event. *Precondition:* Published, provenance-complete resources exist for the concept. *Main flow:* Filter resources by concept and learner preference → select best fit. *Exception:* No qualifying resource → concept flagged rather than substituting an unvetted resource. *Postcondition:* Resource assigned or gap flagged.

**UC-11 — Asking "Why Am I Learning This?"**
*Actor:* Learner. *Trigger:* Learner requests explanation. *Precondition:* A path/resource decision exists. *Main flow:* System reconstructs decision rationale → presents explanation. *Exception:* Rationale cannot be reconstructed → system discloses this rather than fabricating a reason. *Postcondition:* Learner receives an accurate explanation or an honest limitation notice.

**UC-12 — Completing a Project**
*Actor:* Learner. *Trigger:* Concept requires project-level evidence. *Precondition:* Concept resource(s) completed. *Main flow:* Learner submits project artifact → automated and/or curator evaluation → result recorded. *Exception:* Inconclusive automated evaluation → routed to curator review. *Postcondition:* Competency record updated from project evidence.

**UC-13 — Reaching the Goal**
*Actor:* Learner, System. *Trigger:* All required concepts for the goal reach Competent state. *Precondition:* Active path near completion. *Main flow:* System verifies full competency coverage against goal's knowledge structure → marks goal achieved → notifies learner. *Exception:* Partial coverage (e.g., a concept never reached Competent) → goal not marked achieved; remaining gaps surfaced. *Postcondition:* Goal-achieved state recorded, fully evidence-backed.

**UC-14 — Administrative Resource/Knowledge-Structure Management**
*Actor:* Content Curator, Platform Administrator. *Trigger:* New resource candidate, staleness flag, or structural review need. *Precondition:* Curator/admin authenticated with appropriate scope. *Main flow:* Review candidate/flag → approve, edit, or retire → change logged. *Exception:* Structural change would introduce a circular/invalid dependency → rejected with diagnostic detail. *Postcondition:* Knowledge structure/resource catalog updated and auditable.

---

## 7. Edge Cases

| Condition | Expected System Behavior |
|---|---|
| Unknown/unsupported goal | Inform learner; do not auto-generate an unvetted curriculum; flag for curator consideration |
| Unknown skill/concept referenced | Reject or flag reference; do not silently create an ungoverned concept |
| Conflicting expert roadmaps (e.g., overlapping domains with contradictory prerequisites) | Flag conflict for curator resolution; do not auto-merge or arbitrarily pick one |
| Missing resource for a required concept | Surface as a path gap; notify curator; do not substitute an unvetted resource |
| Broken resource (dead link, inaccessible) | Withhold from rotation; trigger curator review; offer fallback resource if available |
| Low-quality resource (poor learner feedback/outcomes) | Flag via aggregated feedback signal (FR-FDBK-01) for curator review; do not auto-remove without review |
| Insufficient learner evidence to determine competency | Leave state at In Progress/Weak Evidence; do not default to Competent |
| Contradictory learner signals (e.g., passes quiz, fails project on same concept) | Weight toward the higher-fidelity evidence type per defined rule; flag for curator visibility if ambiguous |
| Assessment failure (learner does not pass) | Trigger targeted remediation (UC-09); do not block learner from other independent path branches unnecessarily |
| AI/automated grading failure or low-confidence result | Route to curator review rather than force a pass/fail outcome |
| External API/data-source failure | Degrade gracefully (cached/last-known-good data where safe); notify affected users of temporary limitation |
| Database/storage failure | Preserve data integrity of already-committed evidence; fail closed on new writes rather than risk silent data loss |
| Stale knowledge structure or resource (past freshness threshold) | Withhold from active use pending curator re-verification |
| Content temporarily unavailable | Offer fallback resource if one exists; otherwise clearly communicate the temporary gap to the learner |

---

## 8. Traceability (Overview)

Detailed matrix in Section 16. Model:

`Business Requirement → Product Capability → Functional Requirement → User Flow → Logical Component`

Example chain: *"Competency must be evidence-based, not completion-based"* (blueprint principle) → *Competency Evaluation capability* → **FR-COMP-01, FR-PROG-01** → UC-07, UC-08 → *Competency Engine (logical)*.

---

## 9. Data Requirements (Conceptual — No Schema)

Key conceptual entities and their essential relationships:

- **Learner Profile:** identity, constraints, preferences, declared prior experience — owned by one Learner.
- **Goal:** learner-defined objective, linked to one Knowledge Structure.
- **Knowledge Structure / Concept:** domain, concepts, prerequisite relationships, versioned; owned/maintained by Curators.
- **Resource:** metadata (source, provenance, format, freshness status), linked to one or more Concepts.
- **Assessment Definition:** per-concept assessment format and scoring criteria.
- **Evidence Record:** a learner's submitted response/artifact and its score, linked to a Learner, a Concept, and an Assessment Definition.
- **Competency Record:** per-learner, per-concept state, always linked to the Evidence Record(s) that justify its current value.
- **Path Instance:** a learner's ordered sequence of concept/resource assignments, derived from Gap analysis and Knowledge Structure constraints.
- **Notification:** event-linked message to a Learner.
- **Feedback Record:** learner rating/flag linked to a Resource or Path decision.
- **Audit Log Entry:** actor, action, target entity, and timestamp for curator/admin changes.

---

## 10. Integration Requirements

- Integration with external resource sources (for discovery) sufficient to capture provenance metadata at ingestion.
- Optional identity/SSO integration for authentication (MVP may use direct credential auth only).
- Notification delivery integration (e.g., email; push/SMS as later scope).
- Integration point for automated assessment/grading logic, with a defined escalation path to human curator review when confidence is low.
- Analytics/logging integration sufficient to support Observability and Audit non-functional targets.

*(Specific vendors/protocols are an architecture-stage decision, not defined here.)*

---

## 11. Security Requirements

- All authentication credentials stored using industry-standard hashing; no plaintext credential storage.
- Role-based access control enforced for every action defined in Section 3, matching the permissions in Section 2.
- Learner data isolated per-account; no cross-learner data access without explicit administrative/support authorization and audit logging.
- All data in transit encrypted; data at rest encrypted for sensitive fields at minimum (credentials, personal identifiers).
- Rate limiting/anti-abuse controls on authentication and assessment-submission endpoints.
- Assessment integrity controls to reduce trivial gaming of evidence (e.g., detecting implausible response patterns) — specific mechanisms deferred to design stage, but the requirement itself is in-scope for SRS.

---

## 12. Error Handling Requirements

- All user-facing errors shall be clear and non-revealing of sensitive internals (e.g., authentication errors do not confirm whether an email is registered).
- All external dependency calls (resource fetch, grading service, notification delivery) shall have defined retry/backoff behavior and a graceful degradation path.
- Failures that would otherwise silently corrupt competency data shall fail closed (reject the write, log the failure) rather than fail open.
- System-level errors shall be logged with enough context (without exposing sensitive learner data in logs) to support diagnosis.

---

## 13. Audit Requirements

- All Curator and Administrator actions affecting knowledge structures, resources, roles, or configuration shall be logged with actor identity and timestamp (BR-12).
- All competency state changes shall be logged with a reference to the originating evidence (BR-08).
- Audit logs shall be retained per a defined policy (specific retention period to be set at the governance/architecture stage) and shall not be editable by the actors they record.

---

## 14. Constraints

- The system depends on the availability of expert-vetted knowledge structures; features cannot function for a domain without one.
- The system depends on the continued availability/stability of externally sourced resources; broken/stale resources are expected and must be handled per Section 7.
- MVP scope is intentionally limited to a small number of domains, per the approved blueprint.
- The system must operate within the freemium business model's tier constraints without misrepresenting learner competency (BR-11).

---

## 15. Assumptions

- Expert-authored or expert-vetted knowledge structures can be sourced or built for the initial target domain(s) within project timelines.
- A sufficient supply of high-quality, curatable external resources exists per concept in the initial domain(s).
- Assessment formats defined for MVP (quiz, at least one applied format) are sufficient to produce meaningful competency evidence for the initial domain(s).
- Learners will accept assessment-based gating in exchange for a more trustworthy progress signal, consistent with the blueprint's value proposition.

---

## 16. Out-of-Scope (v1)

- Full in-house content authoring platform/CMS.
- Certification or credential issuance.
- Social, community, or peer-accountability features.
- Expert marketplace functionality.
- Recruiter-facing competency profiles/data products.
- Database schema design and specific cloud/infrastructure selection.
- Specific ML/AI algorithm or model selection for grading, discovery, or adaptation.

---

## 17. Requirement Traceability Matrix

| Business Requirement (Blueprint) | Product Capability | Functional Requirement(s) | User Flow(s) | Logical Component |
|---|---|---|---|---|
| Learners need an accurate model of what they actually know | Skill Representation & Competency Evaluation | FR-SKILL-01, FR-COMP-01 | UC-07, UC-08 | Competency Engine |
| Curriculum must come from expert structures, not free generation | Expert Roadmap Management | FR-EXPERT-01 | UC-14 | Knowledge Structure Manager |
| Resources must be trusted, provenance-tracked, and fresh | Resource Discovery & Curation | FR-DISC-01, FR-CUR-01 | UC-10, UC-14 | Resource Catalog Manager |
| Path must be personalized within valid expert constraints | Personalized Roadmap Generation | FR-PATH-01 | UC-05 | Path Generation Engine |
| Progress is competency-based, not completion-based | Progress Tracking | FR-PROG-01 | UC-02, UC-13 | Competency Engine |
| Weak areas must trigger targeted remediation | Remediation | FR-REM-01 | UC-09 | Adaptation Engine |
| Path must adapt continuously to new evidence | Adaptive Decisions | FR-ADAPT-01 | UC-08, UC-09 | Adaptation Engine |
| Learner should understand why a decision was made | Explainability | FR-EXPL-01 | UC-11 | Path Generation Engine (rationale output) |
| System must onboard and baseline new learners | Onboarding, Goal Definition, Diagnostic Assessment | FR-ONB-01, FR-GOAL-01, FR-DIAG-01 | UC-01, UC-03, UC-04 | Onboarding Flow |
| Curators/admins must operate with accountability | Administrative Controls | FR-ADMIN-01 | UC-14 | Admin Console |

---

*End of SRS. This document is the authoritative input for the next stage: UI, system architecture, database, and ML design.*
