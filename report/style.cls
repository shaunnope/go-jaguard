% writeup.cls
% Place in \tex\latex\misc within your TeX directory

% --- METADATA -----------------------------------------------------------------
\def\fileversion{0.618}
\def\filedate{2021/03/29}

\NeedsTeXFormat{LaTeX2e}

\ProvidesClass{writeup}[\filedate\space Version \fileversion\space Document class for assignment submissions by @shaunnope, based on the TU-e AGA Homework Template by Applied Geometric Algorithms group, TU Eindhoven]

% --- OPTION PROCESSSING -------------------------------------------------------
\RequirePackage{kvoptions}
\SetupKeyvalOptions{
	family = writeup,
	prefix = writeup@,
	setkeys = \kvsetkeys
}

\DeclareStringOption[01.012]%
{course}
\DeclareStringOption[HW]%
{assignmenttype}
\DeclareStringOption[1]%
{assignment}

\DeclareStringOption[Some One]%
{name}
\DeclareStringOption[100XXXX]%
{studentnumber}

\DeclareStringOption[1]%
{firstexercise}

% pass unrecognized options to `article` class
\DeclareDefaultOption{%
	\PassOptionsToClass{\CurrentOption}{article}%
}

% process all options
\ProcessKeyvalOptions{writeup}

% load parent class with some default options
\LoadClass[a4paper, 11pt]{article}

% --- PACKAGES AND SETUP -------------------------------------------------------

\RequirePackage[
a4paper,
headheight = 22.5pt,
margin = 1in,
tmargin = \dimexpr 1.2in \relax
]{geometry}

% Encoding
%\RequirePackage[utf8]{inputenc}	% UTF-8 characters                                               
%\RequirePackage[T1]{fontenc}    % Tuki ääkkösille (Finnish names don't cause problems)  

% graphics and layout
\RequirePackage{fancyhdr} % for headers and footers
\RequirePackage{graphicx} % for including figures
%\RequirePackage{epstopdf} % add *.eps figures
\RequirePackage{mathpazo} % use Palation font
\RequirePackage{soul} % strikethrough

% Math
\RequirePackage{amsmath} % use AMS math package
\RequirePackage{amssymb} % use AMS symbols
\RequirePackage{amsthm} % for writing proofs

\RequirePackage{commath}
\RequirePackage{textcomp}
\RequirePackage{gensymb}
\RequirePackage{esint} % for oint

\RequirePackage{multicol} % for multicolumns
\RequirePackage[export]{adjustbox} % for aligning figures

%\RequirePackage{esvect}

\RequirePackage{array} % for setting up arguments to columns
\RequirePackage{booktabs} % for professional tables
\RequirePackage
[tworuled, linesnumbered, noend, noline]%
{algorithm2e} % for typesetting pseudo-code
\RequirePackage{xcolor} % for colored text (comments in algorithms)
\RequirePackage{trimspaces, xstring} % for multiple author parsing
\RequirePackage{comment} % for comments

\RequirePackage{enumitem}
\newlist{subsub}{enumerate}{2}
\setlist[subsub]{label=\roman*),leftmargin=4em,itemsep=10pt}

%\RequirePackage[version=4]{mhchem} % for chemical equations and symbols

%\RequirePackage{stmaryrd}
\RequirePackage{enumitem}

\RequirePackage{hyperref}

\graphicspath{{figures/}}

% set up headers and footers; same for all pages
\setlength{\headheight}{26pt}
\newcommand{\writeup@setupHF}{%
	\fancyhf{}
	\fancyhead[L]{}
	\fancyhead[R]{\sffamily \writeup@course \space \writeup@assignmenttype \writeup@assignment \\ \writeup@name\space - \writeup@studentnumber}
	\fancyfoot[L]{}
	\fancyfoot[C]{\thepage}
	\renewcommand{\headrulewidth}{0.5pt}
	\renewcommand{\footrulewidth}{0.4pt}
}
\fancypagestyle{plain}{\writeup@setupHF{}}
\writeup@setupHF{}
\pagestyle{fancy}

% add some keywords
\SetKwInOut{KwReq}{Require}
\SetKwProg{Fn}{function}{:}{}
\SetKw{And}{and}
\SetKw{Or}{or}
\SetKw{From}{from}
\SetKw{To}{to}
\SetKw{Each}{each}
\SetKw{In}{in}
\SetKwArray{arrInt}{new int}
\SetKwFunction{algInc}{++}

% Function style
\SetFuncSty{textsc}
% make it possible to not have line numbers on a single line
\let\oldnl\nl
\newcommand{\nonl}{\renewcommand{\nl}{\let\nl\oldnl}}
% style algorithm captions
\newcommand{\writeup@capsty}[1]{\!}
\SetAlCapSty{writeup@capsty}
\SetAlCapNameFnt{\scshape\hspace{.69em}}
% get rid of semicolons
\DontPrintSemicolon
% comment formatting
\newcommand\agacommfont[1]{\footnotesize\ttfamily\textcolor{gray}{#1}}
\SetCommentSty{agacommfont}

% --- USEFUL MACROS ------------------------------------------------------------
% coloured text
\newcommand*{\blue}[1]{\textcolor{blue}{#1}}
\newcommand*{\red}[1]{\textcolor{red}{#1}}

% number sets
\newcommand*{\R}{{\mathbb R}}
\newcommand*{\N}{{\mathbb N}}
\newcommand*{\Z}{{\mathbb Z}}
\newcommand*{\Q}{{\mathbb Q}}

% unit vectors
\newcommand*{\ihat}{\hat{\imath}}
\newcommand*{\jhat}{\hat{\jmath}}
\newcommand*{\khat}{\hat{k}}
\newcommand*{\rhat}{\hat{r}}
\newcommand*{\thhat}{\hat{\theta}}

% Maxwell Equations
\newcommand{\Qe}{Q_{\text{enc}}}
\newcommand{\eo}{\varepsilon_0}
\newcommand{\Ie}{I_{\text{enc}}}

\newcommand{\ep}{\varepsilon}

% Vector algebra
\newcommand{\mat}[1]{\mathbf{#1}}
\renewcommand{\vec}[1]{\mathbf{#1}}
\newcommand{\vect}[1]{\mathbf{#1}}
\newcommand{\Span}[1]{{\text{span} \left(#1\right)}}
\newcommand{\ColVec}[1][1 \\ 2]{\begin{bmatrix}
		#1
	\end{bmatrix}}

% Row Operations
\newcommand{\rowswap}[2]{#1 \leftrightarrow #2}
\newcommand{\rowscale}[2]{#1 \rightarrow #2}
\newcommand{\RowSwap}[2]{\xrightarrow{#1 \leftrightarrow #2}}
\newcommand{\RowScale}[2]{\xrightarrow{#1 \rightarrow #2}}

\newcommand {\MultiRow}[2][]{\xrightarrow[\substack{#1}]{\substack{#2}}}

% Prob and Stats
\renewcommand{\Pr}{\mathbb{P}}
\newcommand{\E}{\mathbb{E}}
\DeclareMathOperator{\V}{Var}
\DeclareMathOperator{\Cov}{Cov}

\newcommand{\Choose}[2]{{}^{#1}\text{C}_{#2}}
\newcommand{\Permute}[2]{{}^{#1}\text{P}_{#2}}

\newcommand{\stcomp}[1]{{#1}^{\mathsf{c}}}

\DeclareMathOperator{\Bin}{\mathtt{binomial}}
\DeclareMathOperator{\Geo}{\mathtt{geometric}}
\DeclareMathOperator{\Poi}{\mathtt{Poisson}}
\DeclareMathOperator{\Normal}{\mathcal{N}}
\DeclareMathOperator{\Exp}{\mathtt{exponential}}

% cryptography
\newcommand{\xor}{\, \oplus}
\newcommand{\Expt}[2][OW]{\mathsf{Exp}^{\mathsf{#1}}_{#2}}
\newcommand{\ExO}[1]{\Expt[OW]{\mathit{#1}, \Adv_{\mathit{#1}}}}
\newcommand{\Adv}{\mathcal{A}}
\newcommand{\Advt}[2][PRG]{\mathsf{Adv}^{\mathsf{#1}}_{#2}}


% --- CUSTOM COMMANDS ----------------------------------------------------------
\newcounter{question}[section]
\setcounter{question}{\writeup@firstexercise}
\addtocounter{question}{-1}
\newcounter{subquestion}[question]
\newcounter{subsubquestion}[subquestion]
\newcommand*{\question}[1][]{%
	\subsection*{\stepcounter{question}\arabic{question}. #1}%
}

\newcommand*{\subquestion}[1][]{%
	\paragraph{(\ifx/#1/\protect\stepcounter{subquestion}\alph{subquestion}\else#1\fi)}}

\newcommand*{\subsubquestion}[1][]{%
	\paragraph{\ifx/#1/\protect\stepcounter{subsubquestion}\roman{subsubquestion}\else#1\fi)}}

\newenvironment{solution}
{\begin{center}
		\begin{tabular}{|p{0.95\textwidth}|}
			\hline\\
		}
		{ 
			\\ \\ \hline
		\end{tabular} 
	\end{center}
}

\newcommand*{\qrule}{\par\noindent\rule{0.9\textwidth}{0.4pt}}